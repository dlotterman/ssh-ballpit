#!/usr/bin/env python3
"""Inventory the Go module graph and a built OCI image with Syft; write Markdown."""

import argparse
import datetime
import hashlib
import json
import os
import re
from pathlib import Path
import subprocess
import tarfile
import tempfile
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[1]


def run(*args):
    return subprocess.check_output(args, cwd=ROOT, text=True)


def json_stream(text):
    decoder = json.JSONDecoder()
    while text.strip():
        text = text.lstrip()
        item, end = decoder.raw_decode(text)
        yield item
        text = text[end:]


def cell(value):
    return str(value).replace("|", "&#124;").replace("\n", " ")


def licenses(artifact):
    return sorted({
        entry.get("spdxExpression") or entry["value"]
        for entry in artifact.get("licenses", [])
        if not entry.get("value", "").startswith("sha256:")
    })


def voice_inventory(archive_path):
    """Read Codex's supplementary source manifest without executing the image."""
    found = {}
    with tarfile.open(archive_path) as archive:
        index = json.load(archive.extractfile("index.json"))
        if len(index["manifests"]) != 1:
            raise SystemExit("Use a single-platform OCI archive for the inventory")
        digest = index["manifests"][0]["digest"].replace(":", "/")
        manifest = json.load(archive.extractfile("blobs/" + digest))
        for layer in manifest["layers"]:
            blob = archive.extractfile("blobs/" + layer["digest"].replace(":", "/"))
            with tarfile.open(fileobj=blob, mode="r|*") as contents:
                for member in contents:
                    if member.isfile() and member.name.endswith("codex-resources/voice/sources.json"):
                        found[member.name] = json.load(contents.extractfile(member))
    if not found:
        return ["No supplementary Codex voice source manifest was detected in this image.", ""]
    # These labels were checked against the license files shipped with 0.162.1;
    # refresh them when updating Codex, rather than treating Apache-2.0 as blanket coverage.
    terms = {
        "libffi": "MIT; licenses/libffi.txt",
        "pcre2": "BSD-3-Clause WITH PCRE2-exception; bundled sljit: BSD-2-Clause; licenses/PCRE2.md and sljit.txt",
        "zlib": "Zlib; licenses/zlib.txt",
        "opus": "BSD-3-Clause; licenses/Opus.txt",
        "glib": "LGPL-2.1; licenses/LGPL-2.1.txt and NOTICE.md",
        "proxy-libintl": "LGPL-2.0; licenses/proxy-libintl.txt",
        "gstreamer": "LGPL-2.1; licenses/LGPL-2.1.txt and NOTICE.md",
        "gst-plugins-base": "LGPL-2.1; licenses/LGPL-2.1.txt and NOTICE.md",
        "gst-plugins-good": "LGPL-2.1; licenses/LGPL-2.1.txt and NOTICE.md",
    }
    rows = ["## Supplementary Codex native-library inventory", "",
            "Codex ships native voice libraries even though ssh-ballpit does not use voice. "
            "These are easy to miss in package-manager scans. The following versions and source "
            "hashes come from the image's `codex-resources/voice/sources.json`; license labels "
            "were checked against its accompanying `NOTICE.md` and `licenses/` files. LGPL version "
            "labels identify the supplied license texts; consult upstream files for any later-version options.", "",
            "Meson and Ninja entries in that manifest are upstream build inputs, not bundled runtime "
            "libraries, and are excluded from this supplementary runtime table. GVDB is included "
            "with GLib under LGPL; sljit is included with PCRE2 under BSD-2-Clause.", "",
            "| Component | Version / revision | License evidence | Source archive SHA256 |",
            "| --- | --- | --- | --- |"]
    for path, metadata in sorted(found.items()):
        for source in metadata["sources"]:
            if source["role"] != "native-library":
                continue
            if source["name"] not in terms:
                raise SystemExit(f"Review new Codex native component: {source['name']}")
            rows.append("| " + " | ".join(map(cell, [f"[{source['name']}]({source['url']})",
                                                       source["version"], terms[source["name"]], source["sha256"]])) + " |")
        for source in metadata.get("bundled_sources", []):
            rows.append("| " + " | ".join(map(cell, [source["name"], source["revision"],
                                                       "LGPL; included by " + source["included_by"],
                                                       "Included in parent archive"])) + " |")
        rows += ["", f"Image evidence: `{path}` and the adjacent `NOTICE.md`, `manifest.json`, and `licenses/`.", ""]
    return rows


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--image-archive", type=Path, required=True,
                        help="OCI archive from podman save --format oci-archive")
    parser.add_argument("--output", type=Path, default=ROOT / "sbom.md")
    args = parser.parse_args()
    go, syft = os.environ.get("GO", "go"), os.environ.get("SYFT", "syft")
    go_version = re.search(r"^go (.+)$", (ROOT / "go.mod").read_text(), re.MULTILINE).group(1)
    codex_version = re.search(r"^ARG CODEX_VERSION=(.+)$", (ROOT / "Containerfile").read_text(), re.MULTILINE).group(1)
    if codex_version != "0.162.1":
        raise SystemExit("Review and update the supplementary Codex license mapping for the new version")
    # Download exact selected versions and their license files, not new versions.
    subprocess.run([go, "mod", "download", "all"], cwd=ROOT, check=True)
    modules = [m for m in json_stream(run(go, "list", "-m", "-json", "all"))
               if not m.get("Main")]
    if any(m.get("Replace") for m in modules):
        raise SystemExit("Review replacement-module identity and licenses before generating an SBOM")
    imported = {p["Module"]["Path"] for p in json_stream(
        run(go, "list", "-deps", "-json", "./cmd/ssh-ballpit")) if "Module" in p}
    direct = {m["Path"] for m in modules if not m.get("Indirect")}
    with tempfile.TemporaryDirectory(prefix="ssh-ballpit-sbom-") as directory:
        temporary = Path(directory)
        graph = temporary / "graph"
        graph.mkdir()
        # A synthetic manifest lets Syft enrich every selected graph module,
        # including modules not directly required by this project's go.mod.
        (graph / "go.mod").write_text(f"module ssh-ballpit-inventory\n\ngo {go_version}\n\nrequire (\n" +
                                     "".join(f"\t{m['Path']} {m['Version']}\n" for m in modules) + ")\n")
        (graph / "go.sum").write_bytes((ROOT / "go.sum").read_bytes())
        source_path, image_path = temporary / "source.json", temporary / "image.json"
        subprocess.run([syft, "scan", f"dir:{graph}", "--source-name", "ssh-ballpit-module-graph",
                        "--enrich", "golang", "-o", f"syft-json={source_path}"], check=True)
        subprocess.run([syft, "scan", f"oci-archive:{args.image_archive.resolve()}",
                        "--enrich", "golang", "-o", f"syft-json={image_path}"], check=True)
        source, image = json.loads(source_path.read_text()), json.loads(image_path.read_text())

    image_modules = {(a["name"], a["version"]) for a in image["artifacts"] if a["type"] == "go-module"}
    expected = {(m["Path"], m["Version"]) for m in modules if m["Path"] in imported}
    if not expected <= image_modules:
        raise SystemExit(f"Rebuild the image: imported Go modules differ: {sorted(expected - image_modules)}")
    if not any(a["name"] == "@openai/codex" and a["version"] == codex_version for a in image["artifacts"]):
        raise SystemExit("Rebuild the image: Codex version differs from Containerfile")
    node_version = next(a["version"] for a in image["artifacts"] if a["name"] == "node" and a["type"] == "binary")
    npm_version = next(a["version"] for a in image["artifacts"] if a["name"] == "npm" and a["type"] == "npm")

    artifacts = {(a["name"], a["version"]): a for a in source["artifacts"]}
    for module in modules:
        key = module["Path"], module["Version"]
        if key not in artifacts or not licenses(artifacts[key]):
            raise SystemExit(f"Missing module license evidence: {key}")

    lines = ["# Software bill of materials", "",
             f"Generated {datetime.datetime.now(datetime.timezone.utc).date().isoformat()} with "
             f"Syft {source['descriptor']['version']}. Host inventory toolchain: `{run(go, 'version').strip()}`.", "",
             "## Scope and provenance", "",
             "This snapshot covers the exact selected Go module graph, the modules imported by the "
             "Linux executable, and packages detected in the built runtime container. Versions are "
             "observed or selected versions, not recommendations. Go modules are not upgraded by this report.", "",
             f"- Go modules: **{len(modules)}**, of which **{sum(m['Path'] in imported for m in modules)}** "
             "are imported by the executable. The remaining graph modules are not automatically linked into it.",
             f"- `go.mod` SHA256: `{hashlib.sha256((ROOT / 'go.mod').read_bytes()).hexdigest()}`.",
             f"- `go.sum` SHA256: `{hashlib.sha256((ROOT / 'go.sum').read_bytes()).hexdigest()}`.",
             f"- Runtime image ID: `{image['source']['metadata'].get('imageID', 'not reported')}`.",
             f"- Exported OCI manifest digest: `{image['source']['metadata'].get('manifestDigest', 'not reported')}`.",
             f"- Runtime image platform: `{image['source']['metadata'].get('os', 'unknown')}/"
             f"{image['source']['metadata'].get('architecture', 'unknown')}`.", "",
             "The image was built from [Containerfile](Containerfile). Its Go builder and Node base tags "
             "are mutable, so rebuilds can change the toolchain, OS packages and bundled npm versions. "
             "This inventories the final runtime image, not the discarded build image or the host OS. "
             "Exporting to OCI can change the manifest digest while preserving the image config ID.", "",
             "## License interpretation", "",
             "The Unlicense covers only this project's original code and packaging. Third-party code "
             "retains its own licenses and notices. License sets below describe files found in a module "
             "or package; multiple identifiers do not imply a choice of license for every file.", "",
             "Go license evidence comes from the exact downloaded module versions, including nested "
             "LICENSE/COPYING files and NOTICE files, enriched by Syft. Version-specific package links "
             "provide a second way to inspect upstream license information. In particular, "
             "`go.yaml.in/yaml/v3` contains MIT-covered libyaml ports and Apache-2.0-covered files; "
             "`client_golang` contains BSD-3-Clause code alongside Apache-2.0 code.", "",
             "Container license expressions are scanner observations of package metadata and installed "
             "copyright files. `NOASSERTION` means the scanner could not establish a license, not that "
             "the component is unlicensed or covered by The Unlicense. Hash-only NOTICE fingerprints "
             "are not treated as additional license identifiers.", "",
             "Precompiled Codex, Node and other native binaries may contain components without "
             "discoverable package metadata. This package-level SBOM is not a complete file-level "
             "license audit of those binaries. Consult upstream notices and the image's "
             "`/usr/share/doc/*/copyright` files when redistributing. External LLM services, model "
             "weights, host sshd and developer tools are not bundled application imports.", ""]

    for linked, title in [(True, "Go modules imported by the executable"),
                          (False, "Additional selected Go module graph")]:
        lines += [f"## {title}", "",
                  "| Module | Version | Relationship | License identifiers found | Evidence within module |",
                  "| --- | --- | --- | --- | --- |"]
        for module in modules:
            name, version = module["Path"], module["Version"]
            if (name in imported) != linked:
                continue
            artifact = artifacts[name, version]
            prefix = f"file://$GOPATH/pkg/mod/{name}@{version}/"
            evidence = sorted({url[len(prefix):] for license in artifact.get("licenses", [])
                               for url in license.get("urls", []) if url.startswith(prefix)})
            relationship = "Direct" if name in direct else ("Transitive import" if linked else "Graph only")
            url = f"https://pkg.go.dev/{quote(name, safe='/')}@{quote(version)}?tab=licenses"
            lines.append("| " + " | ".join(map(cell, [f"[{name}]({url})", version, relationship,
                                                       "; ".join(licenses(artifact)), "; ".join(evidence)])) + " |")
        lines.append("")

    lines += ["The Go runtime and standard library are also part of the executable (BSD-3-Clause "
              "with bundled third-party notices). See the detected `stdlib` version in the image "
              "inventory and [Go's LICENSE](https://go.dev/LICENSE). Module checksums are in "
              "[go.sum](go.sum).", "", "## Container runtime inventory", "",
              f"The recipe installs `@openai/codex@{codex_version}` (Apache-2.0), Node.js (MIT plus bundled "
              "notices), npm (Artistic-2.0 plus dependencies), and Debian packages. The table below "
              "includes the versions and license metadata actually detected in this image. "
              f"See [Codex's license](https://github.com/openai/codex/blob/rust-v{codex_version}/LICENSE), "
              f"[Node's license and bundled notices](https://github.com/nodejs/node/blob/v{node_version}/LICENSE), "
              f"and [npm's license](https://github.com/npm/cli/blob/v{npm_version}/LICENSE).", ""]
    # Preserve all distinct package/version/license combinations; duplicate
    # installations are represented once, retaining their evidence paths.
    packages = {}
    for artifact in image["artifacts"]:
        terms = "; ".join(licenses(artifact)) or "NOASSERTION"
        if artifact["name"] == "ssh-ballpit" or artifact["name"] == "ssh-ballpit/cmd/ssh-ballpit":
            terms = "Unlicense (project declaration)"
        key = artifact["type"], artifact["name"], artifact["version"], terms
        paths = packages.setdefault(key, set())
        paths.update(location["path"] for location in artifact.get("locations", []))
    lines += [f"Detected **{len(packages)} distinct package/version/license records**. "
              "The Go rows here describe embedded binary metadata, not additional copies of the source modules.", "",
              "<details>", "<summary>Complete detected runtime package list</summary>", "",
              "| Type | Package | Version | License metadata | Evidence path in image |",
              "| --- | --- | --- | --- | --- |"]
    for key, paths in sorted(packages.items()):
        evidence = sorted(path for path in paths if path.endswith(("/copyright", "/package.json")))
        row = [*key, "; ".join(evidence or sorted(paths))]
        lines.append("| " + " | ".join(map(cell, row)) + " |")
    lines += ["", "</details>", ""]
    lines += voice_inventory(args.image_archive)
    lines += ["## Refresh", "",
              "Install Go, Python 3, rootless Podman, and [Syft](https://github.com/anchore/syft). "
              "Run from the repository root; the archive can be stored outside the checkout:", "",
              "```sh", "podman build -t localhost/ssh-ballpit:sbom .",
              "podman save --format oci-archive -o /tmp/ssh-ballpit-sbom.oci localhost/ssh-ballpit:sbom",
              "python3 scripts/generate-sbom.py --image-archive /tmp/ssh-ballpit-sbom.oci", "```", "",
              "`GO` and `SYFT` can select explicit tool paths. Generation downloads the selected "
              "module graph, may add missing checksums to `go.sum`, and uses network access for license "
              "enrichment. Review changed licenses and any `NOASSERTION` entries, and update the README "
              "license summary when dependencies or the container recipe change.", ""]
    args.output.write_text("\n".join(lines))
    print(f"Wrote {args.output}: {len(modules)} Go modules, {len(packages)} runtime records")


if __name__ == "__main__":
    main()
