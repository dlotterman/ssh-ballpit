# Software bill of materials

Generated 2026-10-14 with Syft 1.54.1. Host inventory toolchain: `go version go1.27.2 linux/amd64`.

## Scope and provenance

This snapshot covers the exact selected Go module graph, the modules imported by the Linux executable, and packages detected in the built runtime container. Versions are observed or selected versions, not recommendations. Go modules are not upgraded by this report.

- Go modules: **44**, of which **12** are imported by the executable. The remaining graph modules are not automatically linked into it.
- `go.mod` SHA256: `a3d67bbe0ae423703bcb52cb6001d12d07e64bdd35753079fa747da837648065`.
- `go.sum` SHA256: `6f9b7532480dbb7e86f9fa437f43532ba07c7e889d95dac174d6d8f0b6dba510`.
- Runtime image ID: `sha256:7db3db386865953c6ff574d7ccdfa520fd6df45c31d14826eb860ea73f75bcd5`.
- Exported OCI manifest digest: `sha256:001c0df346c024d54bf630d94851ddb22c6a61b40e5cc43eb81ed7db253ddc6e`.
- Runtime image platform: `linux/amd64`.

The image was built from [Containerfile](Containerfile). Its Go builder and Node base tags are mutable, so rebuilds can change the toolchain, OS packages and bundled npm versions. This inventories the final runtime image, not the discarded build image or the host OS. Exporting to OCI can change the manifest digest while preserving the image config ID.

## License interpretation

The Unlicense covers only this project's original code and packaging. Third-party code retains its own licenses and notices. License sets below describe files found in a module or package; multiple identifiers do not imply a choice of license for every file.

Go license evidence comes from the exact downloaded module versions, including nested LICENSE/COPYING files and NOTICE files, enriched by Syft. Version-specific package links provide a second way to inspect upstream license information. In particular, `go.yaml.in/yaml/v3` contains MIT-covered libyaml ports and Apache-2.0-covered files; `client_golang` contains BSD-3-Clause code alongside Apache-2.0 code.

Container license expressions are scanner observations of package metadata and installed copyright files. `NOASSERTION` means the scanner could not establish a license, not that the component is unlicensed or covered by The Unlicense. Hash-only NOTICE fingerprints are not treated as additional license identifiers.

Precompiled Codex, Node and other native binaries may contain components without discoverable package metadata. This package-level SBOM is not a complete file-level license audit of those binaries. Consult upstream notices and the image's `/usr/share/doc/*/copyright` files when redistributing. External LLM services, model weights, host sshd and developer tools are not bundled application imports.

## Go modules imported by the executable

| Module | Version | Relationship | License identifiers found | Evidence within module |
| --- | --- | --- | --- | --- |
| [github.com/beorn7/perks](https://pkg.go.dev/github.com/beorn7/perks@v1.0.1?tab=licenses) | v1.0.1 | Transitive import | MIT | LICENSE |
| [github.com/cespare/xxhash/v2](https://pkg.go.dev/github.com/cespare/xxhash/v2@v2.3.0?tab=licenses) | v2.3.0 | Transitive import | MIT | LICENSE.txt |
| [github.com/munnerz/goautoneg](https://pkg.go.dev/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822?tab=licenses) | v0.0.0-20191010083416-a7dc8b61c822 | Transitive import | BSD-3-Clause | LICENSE |
| [github.com/prometheus/client_golang](https://pkg.go.dev/github.com/prometheus/client_golang@v1.25.0?tab=licenses) | v1.25.0 | Direct | Apache-2.0; BSD-3-Clause | LICENSE; NOTICE; internal/github.com/golang/gddo/LICENSE |
| [github.com/prometheus/client_model](https://pkg.go.dev/github.com/prometheus/client_model@v0.6.3?tab=licenses) | v0.6.3 | Transitive import | Apache-2.0 | LICENSE; NOTICE |
| [github.com/prometheus/common](https://pkg.go.dev/github.com/prometheus/common@v0.72.0?tab=licenses) | v0.72.0 | Transitive import | Apache-2.0 | LICENSE; NOTICE |
| [github.com/prometheus/procfs](https://pkg.go.dev/github.com/prometheus/procfs@v0.22.0?tab=licenses) | v0.22.0 | Transitive import | Apache-2.0 | LICENSE; NOTICE |
| [go.yaml.in/yaml/v3](https://pkg.go.dev/go.yaml.in/yaml/v3@v3.0.5?tab=licenses) | v3.0.5 | Direct | Apache-2.0; MIT | LICENSE; NOTICE |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto@v0.58.0?tab=licenses) | v0.58.0 | Direct | BSD-3-Clause | LICENSE |
| [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys@v0.49.0?tab=licenses) | v0.49.0 | Transitive import | BSD-3-Clause | LICENSE |
| [google.golang.org/protobuf](https://pkg.go.dev/google.golang.org/protobuf@v1.36.12?tab=licenses) | v1.36.12 | Transitive import | BSD-3-Clause | LICENSE |
| [mvdan.cc/sh/v3](https://pkg.go.dev/mvdan.cc/sh/v3@v3.14.1?tab=licenses) | v3.14.1 | Direct | BSD-3-Clause | LICENSE |

## Additional selected Go module graph

| Module | Version | Relationship | License identifiers found | Evidence within module |
| --- | --- | --- | --- | --- |
| [github.com/alecthomas/kingpin/v2](https://pkg.go.dev/github.com/alecthomas/kingpin/v2@v2.4.0?tab=licenses) | v2.4.0 | Graph only | MIT | COPYING |
| [github.com/alecthomas/units](https://pkg.go.dev/github.com/alecthomas/units@v0.0.0-20240927000941-0f3dac36c52b?tab=licenses) | v0.0.0-20240927000941-0f3dac36c52b | Graph only | MIT | COPYING |
| [github.com/creack/pty](https://pkg.go.dev/github.com/creack/pty@v1.1.24?tab=licenses) | v1.1.24 | Graph only | MIT | LICENSE |
| [github.com/go-quicktest/qt](https://pkg.go.dev/github.com/go-quicktest/qt@v1.102.0?tab=licenses) | v1.102.0 | Graph only | MIT | LICENSE |
| [github.com/golang-jwt/jwt/v5](https://pkg.go.dev/github.com/golang-jwt/jwt/v5@v5.3.1?tab=licenses) | v5.3.1 | Graph only | MIT | LICENSE |
| [github.com/golang/protobuf](https://pkg.go.dev/github.com/golang/protobuf@v1.5.0?tab=licenses) | v1.5.0 | Graph only | BSD-3-Clause | LICENSE |
| [github.com/google/go-cmp](https://pkg.go.dev/github.com/google/go-cmp@v0.7.0?tab=licenses) | v0.7.0 | Graph only | BSD-3-Clause | LICENSE |
| [github.com/google/renameio/v2](https://pkg.go.dev/github.com/google/renameio/v2@v2.0.2?tab=licenses) | v2.0.2 | Graph only | Apache-2.0 | LICENSE |
| [github.com/google/uuid](https://pkg.go.dev/github.com/google/uuid@v1.6.0?tab=licenses) | v1.6.0 | Graph only | BSD-3-Clause | LICENSE |
| [github.com/jpillora/backoff](https://pkg.go.dev/github.com/jpillora/backoff@v1.0.0?tab=licenses) | v1.0.0 | Graph only | MIT | LICENSE |
| [github.com/json-iterator/go](https://pkg.go.dev/github.com/json-iterator/go@v1.1.12?tab=licenses) | v1.1.12 | Graph only | MIT | LICENSE |
| [github.com/julienschmidt/httprouter](https://pkg.go.dev/github.com/julienschmidt/httprouter@v1.3.0?tab=licenses) | v1.3.0 | Graph only | BSD-3-Clause | LICENSE |
| [github.com/klauspost/compress](https://pkg.go.dev/github.com/klauspost/compress@v1.20.1?tab=licenses) | v1.20.1 | Graph only | Apache-2.0; BSD-3-Clause; MIT | LICENSE; gzhttp/LICENSE; internal/lz4ref/LICENSE; internal/snapref/LICENSE; s2/LICENSE; s2/cmd/internal/filepathx/LICENSE; s2/cmd/internal/readahead/LICENSE; snappy/LICENSE; snappy/xerial/LICENSE; zstd/internal/xxhash/LICENSE.txt |
| [github.com/kr/pretty](https://pkg.go.dev/github.com/kr/pretty@v0.3.1?tab=licenses) | v0.3.1 | Graph only | MIT | License |
| [github.com/kr/text](https://pkg.go.dev/github.com/kr/text@v0.2.0?tab=licenses) | v0.2.0 | Graph only | MIT | License |
| [github.com/kylelemons/godebug](https://pkg.go.dev/github.com/kylelemons/godebug@v1.1.0?tab=licenses) | v1.1.0 | Graph only | Apache-2.0 | LICENSE |
| [github.com/modern-go/concurrent](https://pkg.go.dev/github.com/modern-go/concurrent@v0.0.0-20180306012644-bacd9c7ef1dd?tab=licenses) | v0.0.0-20180306012644-bacd9c7ef1dd | Graph only | Apache-2.0 | LICENSE |
| [github.com/modern-go/reflect2](https://pkg.go.dev/github.com/modern-go/reflect2@v1.0.2?tab=licenses) | v1.0.2 | Graph only | Apache-2.0 | LICENSE |
| [github.com/mwitkow/go-conntrack](https://pkg.go.dev/github.com/mwitkow/go-conntrack@v0.0.0-20190716064945-2f068394615f?tab=licenses) | v0.0.0-20190716064945-2f068394615f | Graph only | Apache-2.0 | LICENSE |
| [github.com/rogpeppe/go-internal](https://pkg.go.dev/github.com/rogpeppe/go-internal@v1.15.0?tab=licenses) | v1.15.0 | Graph only | BSD-3-Clause | LICENSE |
| [github.com/stretchr/testify](https://pkg.go.dev/github.com/stretchr/testify@v1.12.1?tab=licenses) | v1.12.1 | Graph only | BSD-3-Clause; ISC; MIT | LICENSE; internal/difflib/LICENSE; internal/spew/LICENSE |
| [github.com/xhit/go-str2duration/v2](https://pkg.go.dev/github.com/xhit/go-str2duration/v2@v2.1.0?tab=licenses) | v2.1.0 | Graph only | BSD-3-Clause | LICENSE |
| [go.uber.org/goleak](https://pkg.go.dev/go.uber.org/goleak@v1.3.0?tab=licenses) | v1.3.0 | Graph only | MIT | LICENSE |
| [go.yaml.in/yaml/v2](https://pkg.go.dev/go.yaml.in/yaml/v2@v2.4.4?tab=licenses) | v2.4.4 | Graph only | Apache-2.0; MIT | LICENSE; LICENSE.libyaml; NOTICE |
| [golang.org/x/mod](https://pkg.go.dev/golang.org/x/mod@v0.29.0?tab=licenses) | v0.29.0 | Graph only | BSD-3-Clause | LICENSE |
| [golang.org/x/net](https://pkg.go.dev/golang.org/x/net@v0.60.0?tab=licenses) | v0.60.0 | Graph only | BSD-3-Clause | LICENSE |
| [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2@v0.37.0?tab=licenses) | v0.37.0 | Graph only | BSD-3-Clause | LICENSE |
| [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync@v0.22.0?tab=licenses) | v0.22.0 | Graph only | BSD-3-Clause | LICENSE |
| [golang.org/x/term](https://pkg.go.dev/golang.org/x/term@v0.47.0?tab=licenses) | v0.47.0 | Graph only | BSD-3-Clause | LICENSE |
| [golang.org/x/text](https://pkg.go.dev/golang.org/x/text@v0.43.0?tab=licenses) | v0.43.0 | Graph only | BSD-3-Clause | LICENSE |
| [golang.org/x/tools](https://pkg.go.dev/golang.org/x/tools@v0.38.0?tab=licenses) | v0.38.0 | Graph only | BSD-3-Clause | LICENSE |
| [mvdan.cc/editorconfig](https://pkg.go.dev/mvdan.cc/editorconfig@v0.3.0?tab=licenses) | v0.3.0 | Graph only | BSD-3-Clause | LICENSE |

The Go runtime and standard library are also part of the executable (BSD-3-Clause with bundled third-party notices). See the detected `stdlib` version in the image inventory and [Go's LICENSE](https://go.dev/LICENSE). Module checksums are in [go.sum](go.sum).

## Container runtime inventory

The recipe installs `@openai/codex@0.162.1` (Apache-2.0), Node.js (MIT plus bundled notices), npm (Artistic-2.0 plus dependencies), and Debian packages. The table below includes the versions and license metadata actually detected in this image. See [Codex's license](https://github.com/openai/codex/blob/rust-v0.162.1/LICENSE), [Node's license and bundled notices](https://github.com/nodejs/node/blob/v22.23.3/LICENSE), and [npm's license](https://github.com/npm/cli/blob/v10.9.9/LICENSE).

Detected **295 distinct package/version/license records**. The Go rows here describe embedded binary metadata, not additional copies of the source modules.

<details>
<summary>Complete detected runtime package list</summary>

| Type | Package | Version | License metadata | Evidence path in image |
| --- | --- | --- | --- | --- |
| binary | node | 22.23.3 | NOASSERTION | /usr/local/bin/node |
| deb | adduser | 3.134 | GPL-2.0-only; GPL-2.0-or-later | /usr/share/doc/adduser/copyright |
| deb | apt | 2.6.1 | BSD-3-Clause; Expat; GPL-2.0-only; GPL-2.0-or-later | /usr/share/doc/apt/copyright |
| deb | base-files | 12.4+deb12u15 | NOASSERTION | /usr/share/doc/base-files/copyright |
| deb | base-passwd | 3.6.1 | GPL-2.0-only; public-domain | /usr/share/doc/base-passwd/copyright |
| deb | bash | 5.2.15-2+b13 | BSD-4-Clause-UC; GFDL-1.3-only; GFDL-NIV-1.3; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; Latex2e; MIT-like; permissive | /usr/share/doc/bash/copyright |
| deb | bsdutils | 1:2.38.1-5+deb12u3 | BSD-3-Clause; BSD-4-Clause; BSLA; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT; public-domain | /usr/share/doc/bsdutils/copyright |
| deb | ca-certificates | 20250419~deb12u1 | GPL-2.0-only; GPL-2.0-or-later; MPL-2.0 | /usr/share/doc/ca-certificates/copyright |
| deb | coreutils | 9.1-1 | BSD-4-Clause-UC; FSFULLR; GFDL-1.3-only; GFDL-NIV-1.3; GPL-3.0-only; GPL-3.0-or-later; ISC | /usr/share/doc/coreutils/copyright |
| deb | dash | 0.5.12-2 | BSD-3-Clause; GPL-2.0-only; GPL-2.0-or-later; public-domain | /usr/share/doc/dash/copyright |
| deb | debconf | 1.5.82 | BSD-2-Clause | /usr/share/doc/debconf/copyright |
| deb | debian-archive-keyring | 2023.3+deb12u2 | NOASSERTION | /usr/share/doc/debian-archive-keyring/copyright |
| deb | debianutils | 5.7-0.5~deb12u1 | GPL-2.0-only; GPL-2.0-or-later; SMAIL-GPL; public-domain | /usr/share/doc/debianutils/copyright |
| deb | diffutils | 1:3.8-4 | FSFAP; FSFULLR; GFDL-1.3-only; GFDL-NIV-1.3; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; X11; public-domain | /usr/share/doc/diffutils/copyright |
| deb | dpkg | 1.21.23 | BSD-2-Clause; GPL-2.0-only; GPL-2.0-or-later; public-domain-s-s-d | /usr/share/doc/dpkg/copyright |
| deb | e2fsprogs | 1.47.0-2+b2 | Apache-2.0; BSD-3-Clause; GPL; GPL-2.0-only; GPL-2.0-or-later; ISC; Kazlib; LGPL-2.0-only; Latex2e | /usr/share/doc/e2fsprogs/copyright |
| deb | findutils | 4.9.0-4 | BSD-3-Clause; FSFAP; FSFULLR; GFDL-1.3-only; GFDL-NIV-1.3+; GPL; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; ISC; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; X11; public-domain | /usr/share/doc/findutils/copyright |
| deb | gcc-12-base | 12.2.0-14+deb12u1 | NOASSERTION | /usr/share/doc/gcc-12-base/copyright |
| deb | gpgv | 2.2.40-1.1+deb12u2 | BSD-3-Clause; CC0-1.0; Expat; GPL-3.0-only; GPL-3.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; RFC-Reference; TinySCHEME; permissive | /usr/share/doc/gpgv/copyright |
| deb | grep | 3.8-5 | GPL-3.0-only; GPL-3.0-or-later | /usr/share/doc/grep/copyright |
| deb | gzip | 1.12-1 | FSF-manpages; GFDL-1.3+-no-invariant; GFDL-3; GPL-3.0-only; GPL-3.0-or-later | /usr/share/doc/gzip/copyright |
| deb | hostname | 3.23+nmu1 | GPL-2.0-only | /usr/share/doc/hostname/copyright |
| deb | init-system-helpers | 1.65.2+deb12u1 | BSD-3-Clause; GPL-2.0-only; GPL-2.0-or-later | /usr/share/doc/init-system-helpers/copyright |
| deb | libacl1 | 2.3.1-3 | GPL-2.0-only; GPL-2.0-or-later; LGPL-2.0-or-later; LGPL-2.1-only | /usr/share/doc/libacl1/copyright |
| deb | libapt-pkg6.0 | 2.6.1 | BSD-3-Clause; Expat; GPL-2.0-only; GPL-2.0-or-later | /usr/share/doc/libapt-pkg6.0/copyright |
| deb | libattr1 | 1:2.5.1-4 | GPL-2.0-only; GPL-2.0-or-later; LGPL-2.0-or-later; LGPL-2.1-only | /usr/share/doc/libattr1/copyright |
| deb | libaudit-common | 1:3.0.9-1 | GPL-1.0-only; GPL-2.0-only; LGPL-2.1-only | /usr/share/doc/libaudit-common/copyright |
| deb | libaudit1 | 1:3.0.9-1 | GPL-1.0-only; GPL-2.0-only; LGPL-2.1-only | /usr/share/doc/libaudit1/copyright |
| deb | libblkid1 | 2.38.1-5+deb12u3 | BSD-3-Clause; BSD-4-Clause; BSLA; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT; public-domain | /usr/share/doc/libblkid1/copyright |
| deb | libbz2-1.0 | 1.0.8-5+b1 | BSD-variant; GPL-2.0-only | /usr/share/doc/libbz2-1.0/copyright |
| deb | libc-bin | 2.36-9+deb12u14 | NOASSERTION | /usr/share/doc/libc-bin/copyright |
| deb | libc6 | 2.36-9+deb12u14 | NOASSERTION | /usr/share/doc/libc6/copyright |
| deb | libcap-ng0 | 0.8.3-1+b3 | GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; LGPL-2.1-only; LGPL-2.1-or-later | /usr/share/doc/libcap-ng0/copyright |
| deb | libcap2 | 1:2.66-4+deb12u3+b1 | BSD-3-Clause; GPL-2.0-only; GPL-2.0-or-later | /usr/share/doc/libcap2/copyright |
| deb | libcom-err2 | 1.47.0-2+b2 | Apache-2.0; BSD-3-Clause; GPL; GPL-2.0-only; GPL-2.0-or-later; ISC; Kazlib; LGPL-2.0-only; Latex2e | /usr/share/doc/libcom-err2/copyright |
| deb | libcrypt1 | 1:4.4.33-2 | NOASSERTION | /usr/share/doc/libcrypt1/copyright |
| deb | libdb5.3 | 5.3.28+dfsg2-1 | Artistic; BSD-3-Clause; BSD-3-clause-fjord; GPL; GPL-3.0-only; MIT-old; MS-PL; Sleepycat; TCL-like; X11; Zlib | /usr/share/doc/libdb5.3/copyright |
| deb | libdebconfclient0 | 0.270 | BSD-2-Clause; GPL-2.0-only; GPL-2.0-or-later | /usr/share/doc/libdebconfclient0/copyright |
| deb | libext2fs2 | 1.47.0-2+b2 | Apache-2.0; BSD-3-Clause; GPL; GPL-2.0-only; GPL-2.0-or-later; ISC; Kazlib; LGPL-2.0-only; Latex2e | /usr/share/doc/libext2fs2/copyright |
| deb | libffi8 | 3.4.4-1 | Expat; GPL; GPL-2.0-or-later; GPL-3.0-or-later; MPL-1.1; X11; public-domain | /usr/share/doc/libffi8/copyright |
| deb | libgcc-s1 | 12.2.0-14+deb12u1 | NOASSERTION | /usr/share/doc/gcc-12-base/copyright |
| deb | libgcrypt20 | 1.10.1-3+deb12u1 | NOASSERTION | /usr/share/doc/libgcrypt20/copyright |
| deb | libgmp10 | 2:6.2.1+dfsg1-1.1 | GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-3.0-only; LGPL-3.0-or-later | /usr/share/doc/libgmp10/copyright |
| deb | libgnutls30 | 3.7.9-2+deb12u7 | NOASSERTION | /usr/share/doc/libgnutls30/copyright |
| deb | libgpg-error0 | 1.46-1 | BSD-3-Clause; GPL-3.0-only; GPL-3.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; g10-permissive | /usr/share/doc/libgpg-error0/copyright |
| deb | libhogweed6 | 3.8.1-2 | Expat; GAP; GPL; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-3.0-or-later; public-domain | /usr/share/doc/libhogweed6/copyright |
| deb | libidn2-0 | 2.3.3-1+b1 | GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-3.0-only; LGPL-3.0-or-later; Unicode | /usr/share/doc/libidn2-0/copyright |
| deb | liblz4-1 | 1.9.4-1 | BSD-2-Clause; GPL-2.0-only; GPL-2.0-or-later | /usr/share/doc/liblz4-1/copyright |
| deb | liblzma5 | 5.4.1-1+deb12u2 | Autoconf; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; LGPL-2.0-only; LGPL-2.1-only; LGPL-2.1-or-later; PD; PD-debian; config-h; noderivs; permissive-fsf; permissive-nowarranty; probably-PD | /usr/share/doc/liblzma5/copyright |
| deb | libmd0 | 1.0.4-2 | BSD-2-Clause; BSD-3-Clause; BSD-3-clause-Aaron-D-Gifford; Beerware; ISC; public-domain-md4; public-domain-md5; public-domain-sha1 | /usr/share/doc/libmd0/copyright |
| deb | libmount1 | 2.38.1-5+deb12u3 | BSD-3-Clause; BSD-4-Clause; BSLA; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT; public-domain | /usr/share/doc/libmount1/copyright |
| deb | libnettle8 | 3.8.1-2 | Expat; GAP; GPL; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-3.0-or-later; public-domain | /usr/share/doc/libnettle8/copyright |
| deb | libp11-kit0 | 0.24.1-2 | Apache-2.0; BSD-3-Clause; ISC; ISC+IBM; LGPL-2.1-only; LGPL-2.1-or-later; permissive-like-automake-output; same-as-rest-of-p11kit | /usr/share/doc/libp11-kit0/copyright |
| deb | libpam-modules | 1.5.2-6+deb12u2 | BSD-3-Clause; BSD-tcp_wrappers; Beerware; GPL; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-2.0-only; LGPL-2.0-or-later; public-domain | /usr/share/doc/libpam-modules/copyright |
| deb | libpam-modules-bin | 1.5.2-6+deb12u2 | BSD-3-Clause; BSD-tcp_wrappers; Beerware; GPL; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-2.0-only; LGPL-2.0-or-later; public-domain | /usr/share/doc/libpam-modules-bin/copyright |
| deb | libpam-runtime | 1.5.2-6+deb12u2 | BSD-3-Clause; BSD-tcp_wrappers; Beerware; GPL; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-2.0-only; LGPL-2.0-or-later; public-domain | /usr/share/doc/libpam-runtime/copyright |
| deb | libpam0g | 1.5.2-6+deb12u2 | BSD-3-Clause; BSD-tcp_wrappers; Beerware; GPL; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-2.0-only; LGPL-2.0-or-later; public-domain | /usr/share/doc/libpam0g/copyright |
| deb | libpcre2-8-0 | 10.42-1+deb12u2 | BSD-2-Clause; BSD-3-Clause; BSD-3-clause-Cambridge; X11; public-domain | /usr/share/doc/libpcre2-8-0/copyright |
| deb | libseccomp2 | 2.5.4-1+deb12u1 | LGPL-2.1-only | /usr/share/doc/libseccomp2/copyright |
| deb | libselinux1 | 3.4-1+b6 | NOASSERTION | /usr/share/doc/libselinux1/copyright |
| deb | libsemanage-common | 3.4-1 | NOASSERTION | /usr/share/doc/libsemanage-common/copyright |
| deb | libsemanage2 | 3.4-1+b5 | NOASSERTION | /usr/share/doc/libsemanage2/copyright |
| deb | libsepol2 | 3.4-2.1 | GPL-2.0-only; GPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; Zlib | /usr/share/doc/libsepol2/copyright |
| deb | libsmartcols1 | 2.38.1-5+deb12u3 | BSD-3-Clause; BSD-4-Clause; BSLA; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT; public-domain | /usr/share/doc/libsmartcols1/copyright |
| deb | libss2 | 1.47.0-2+b2 | Apache-2.0; BSD-3-Clause; GPL; GPL-2.0-only; GPL-2.0-or-later; ISC; Kazlib; LGPL-2.0-only; Latex2e | /usr/share/doc/libss2/copyright |
| deb | libssl3 | 3.0.22-1~deb12u1 | Apache-2.0; Artistic; GPL-1.0-only; GPL-1.0-or-later | /usr/share/doc/libssl3/copyright |
| deb | libstdc++6 | 12.2.0-14+deb12u1 | NOASSERTION | /usr/share/doc/gcc-12-base/copyright |
| deb | libsystemd0 | 252.39-1~deb12u2 | CC0-1.0; Expat; GPL-2.0-only; GPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; public-domain | /usr/share/doc/libsystemd0/copyright |
| deb | libtasn1-6 | 4.19.0-2+deb12u1 | NOASSERTION | /usr/share/doc/libtasn1-6/copyright |
| deb | libtinfo6 | 6.4-4 | BSD-3-Clause; MIT/X11; X11 | /usr/share/doc/libtinfo6/copyright |
| deb | libudev1 | 252.39-1~deb12u2 | CC0-1.0; Expat; GPL-2.0-only; GPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; public-domain | /usr/share/doc/libudev1/copyright |
| deb | libunistring2 | 1.0-2 | FreeSoftware; GFDL-1.2-only; GFDL-1.2-or-later; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT | /usr/share/doc/libunistring2/copyright |
| deb | libuuid1 | 2.38.1-5+deb12u3 | BSD-3-Clause; BSD-4-Clause; BSLA; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT; public-domain | /usr/share/doc/libuuid1/copyright |
| deb | libxxhash0 | 0.8.1-1 | BSD-2-Clause; GPL-2.0-only | /usr/share/doc/libxxhash0/copyright |
| deb | libzstd1 | 1.5.4+dfsg2-5 | BSD-3-Clause; Expat; GPL-2.0-only; Zlib | /usr/share/doc/libzstd1/copyright |
| deb | login | 1:4.13+dfsg1-1+deb12u2 | BSD-3-Clause; GPL-1.0-only; GPL-2.0-only; GPL-2.0-or-later; public-domain | /usr/share/doc/login/copyright |
| deb | logsave | 1.47.0-2+b2 | Apache-2.0; BSD-3-Clause; GPL; GPL-2.0-only; GPL-2.0-or-later; ISC; Kazlib; LGPL-2.0-only; Latex2e | /usr/share/doc/logsave/copyright |
| deb | mawk | 1.3.4.20200120-3.1 | CC-BY-3.0; GPL-2.0-only; X11 | /usr/share/doc/mawk/copyright |
| deb | mount | 2.38.1-5+deb12u3 | BSD-3-Clause; BSD-4-Clause; BSLA; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT; public-domain | /usr/share/doc/mount/copyright |
| deb | ncurses-base | 6.4-4 | BSD-3-Clause; MIT/X11; X11 | /usr/share/doc/ncurses-base/copyright |
| deb | ncurses-bin | 6.4-4 | BSD-3-Clause; MIT/X11; X11 | /usr/share/doc/ncurses-bin/copyright |
| deb | openssl | 3.0.22-1~deb12u1 | Apache-2.0; Artistic; GPL-1.0-only; GPL-1.0-or-later | /usr/share/doc/openssl/copyright |
| deb | passwd | 1:4.13+dfsg1-1+deb12u2 | BSD-3-Clause; GPL-1.0-only; GPL-2.0-only; GPL-2.0-or-later; public-domain | /usr/share/doc/passwd/copyright |
| deb | perl-base | 5.36.0-7+deb12u3 | Artistic; Artistic-2.0; Artistic-dist; BSD-3-Clause; BSD-3-clause-GENERIC; BSD-3-clause-with-weird-numbering; BSD-4-clause-POWERDOG; BZIP; DONT-CHANGE-THE-GPL; Expat; GPL-1.0-only; GPL-1.0-or-later; GPL-2.0-only; GPL-2.0-or-later; GPL-3+-WITH-BISON-EXCEPTION; HSIEH-BSD; HSIEH-DERIVATIVE; LGPL-2.1-only; REGCOMP; REGCOMP,; RRA-KEEP-THIS-NOTICE; SDBM-PUBLIC-DOMAIN; TEXT-TABS; Unicode; Zlib | /usr/share/doc/perl-base/copyright |
| deb | sed | 4.9-1+deb12u1 | BSD-4-Clause-UC; BSL-1.0; GFDL-1.3-only; GFDL-NIV-1.3+; GPL-3.0-only; GPL-3.0-or-later; ISC; X11; pcre | /usr/share/doc/sed/copyright |
| deb | sysvinit-utils | 3.06-4 | GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only | /usr/share/doc/sysvinit-utils/copyright |
| deb | tar | 1.34+dfsg-1.2+deb12u1 | GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL-3.0-only; LGPL-3.0-or-later | /usr/share/doc/tar/copyright |
| deb | tzdata | 2026c-0+deb12u1 | public-domain | /usr/share/doc/tzdata/copyright |
| deb | usr-is-merged | 37~deb12u1 | GPL-2.0-only; GPL-2.0-or-later | /usr/share/doc/usr-is-merged/copyright |
| deb | util-linux | 2.38.1-5+deb12u3 | BSD-3-Clause; BSD-4-Clause; BSLA; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT; public-domain | /usr/share/doc/util-linux/copyright |
| deb | util-linux-extra | 2.38.1-5+deb12u3 | BSD-3-Clause; BSD-4-Clause; BSLA; GPL-2.0-only; GPL-2.0-or-later; GPL-3.0-only; GPL-3.0-or-later; LGPL; LGPL-2.0-only; LGPL-2.0-or-later; LGPL-2.1-only; LGPL-2.1-or-later; LGPL-3.0-only; LGPL-3.0-or-later; MIT; public-domain | /usr/share/doc/util-linux-extra/copyright |
| deb | zlib1g | 1:1.2.13.dfsg-1 | Zlib | /usr/share/doc/zlib1g/copyright |
| go-module | github.com/beorn7/perks | v1.0.1 | MIT | /usr/local/bin/ssh-ballpit |
| go-module | github.com/cespare/xxhash/v2 | v2.3.0 | MIT | /usr/local/bin/ssh-ballpit |
| go-module | github.com/munnerz/goautoneg | v0.0.0-20191010083416-a7dc8b61c822 | BSD-3-Clause | /usr/local/bin/ssh-ballpit |
| go-module | github.com/prometheus/client_golang | v1.25.0 | Apache-2.0; BSD-3-Clause | /usr/local/bin/ssh-ballpit |
| go-module | github.com/prometheus/client_model | v0.6.3 | Apache-2.0 | /usr/local/bin/ssh-ballpit |
| go-module | github.com/prometheus/common | v0.72.0 | Apache-2.0 | /usr/local/bin/ssh-ballpit |
| go-module | github.com/prometheus/procfs | v0.22.0 | Apache-2.0 | /usr/local/bin/ssh-ballpit |
| go-module | go.yaml.in/yaml/v3 | v3.0.5 | Apache-2.0; MIT | /usr/local/bin/ssh-ballpit |
| go-module | golang.org/x/crypto | v0.58.0 | BSD-3-Clause | /usr/local/bin/ssh-ballpit |
| go-module | golang.org/x/sys | v0.49.0 | BSD-3-Clause | /usr/local/bin/ssh-ballpit |
| go-module | google.golang.org/protobuf | v1.36.12 | BSD-3-Clause | /usr/local/bin/ssh-ballpit |
| go-module | mvdan.cc/sh/v3 | v3.14.1 | BSD-3-Clause | /usr/local/bin/ssh-ballpit |
| go-module | ssh-ballpit | UNKNOWN | Unlicense (project declaration) | /usr/local/bin/ssh-ballpit |
| go-module | stdlib | go1.26.9 | BSD-3-Clause | /usr/local/bin/ssh-ballpit |
| npm | @isaacs/cliui | 8.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/@isaacs/cliui/package.json |
| npm | @isaacs/fs-minipass | 4.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/@isaacs/fs-minipass/package.json |
| npm | @isaacs/string-locale-compare | 1.1.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@isaacs/string-locale-compare/package.json |
| npm | @npmcli/agent | 3.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/agent/package.json |
| npm | @npmcli/arborist | 8.0.5 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/arborist/package.json |
| npm | @npmcli/config | 9.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/config/package.json |
| npm | @npmcli/fs | 4.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/fs/package.json |
| npm | @npmcli/git | 6.0.3 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/git/package.json |
| npm | @npmcli/installed-package-contents | 3.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/installed-package-contents/package.json |
| npm | @npmcli/map-workspaces | 4.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/map-workspaces/package.json |
| npm | @npmcli/metavuln-calculator | 8.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/metavuln-calculator/package.json |
| npm | @npmcli/name-from-folder | 3.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/name-from-folder/package.json |
| npm | @npmcli/node-gyp | 4.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/node-gyp/package.json |
| npm | @npmcli/package-json | 6.2.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/package-json/package.json |
| npm | @npmcli/promise-spawn | 8.0.3 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/promise-spawn/package.json |
| npm | @npmcli/query | 4.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/query/package.json |
| npm | @npmcli/redact | 3.2.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/redact/package.json |
| npm | @npmcli/run-script | 9.1.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/run-script/package.json |
| npm | @openai/codex | 0.162.1 | Apache-2.0 | /usr/local/lib/node_modules/@openai/codex/package.json |
| npm | @openai/codex | 0.162.1-linux-x64 | Apache-2.0 | /usr/local/lib/node_modules/@openai/codex/node_modules/@openai/codex-linux-x64/package.json |
| npm | @pkgjs/parseargs | 0.11.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/@pkgjs/parseargs/package.json |
| npm | @sigstore/bundle | 3.1.0 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/@sigstore/bundle/package.json |
| npm | @sigstore/core | 2.0.0 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/@sigstore/core/package.json |
| npm | @sigstore/protobuf-specs | 0.4.3 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/@sigstore/protobuf-specs/package.json |
| npm | @sigstore/sign | 3.1.0 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/@sigstore/sign/package.json |
| npm | @sigstore/tuf | 3.1.1 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/@sigstore/tuf/package.json |
| npm | @sigstore/verify | 2.1.1 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/@sigstore/verify/package.json |
| npm | @tufjs/canonical-json | 2.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/@tufjs/canonical-json/package.json |
| npm | @tufjs/models | 3.0.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/tuf-js/node_modules/@tufjs/models/package.json |
| npm | abbrev | 3.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/abbrev/package.json |
| npm | agent-base | 7.1.4 | MIT | /usr/local/lib/node_modules/npm/node_modules/agent-base/package.json |
| npm | ansi-regex | 5.0.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/ansi-regex/package.json |
| npm | ansi-regex | 6.2.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/@isaacs/cliui/node_modules/ansi-regex/package.json; /usr/local/lib/node_modules/npm/node_modules/wrap-ansi/node_modules/ansi-regex/package.json |
| npm | ansi-styles | 4.3.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/wrap-ansi-cjs/node_modules/ansi-styles/package.json |
| npm | ansi-styles | 6.2.3 | MIT | /usr/local/lib/node_modules/npm/node_modules/ansi-styles/package.json |
| npm | aproba | 2.1.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/aproba/package.json |
| npm | archy | 1.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/archy/package.json |
| npm | balanced-match | 1.0.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/balanced-match/package.json |
| npm | bin-links | 5.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/bin-links/package.json |
| npm | binary-extensions | 2.3.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/binary-extensions/package.json |
| npm | brace-expansion | 2.0.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/brace-expansion/package.json |
| npm | cacache | 19.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/cacache/package.json |
| npm | chalk | 5.6.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/chalk/package.json |
| npm | chownr | 3.0.0 | BlueOak-1.0.0 | /usr/local/lib/node_modules/npm/node_modules/chownr/package.json |
| npm | ci-info | 4.4.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/ci-info/package.json |
| npm | cidr-regex | 4.1.3 | BSD-2-Clause | /usr/local/lib/node_modules/npm/node_modules/cidr-regex/package.json |
| npm | cli-columns | 4.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/cli-columns/package.json |
| npm | cmd-shim | 7.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/cmd-shim/package.json |
| npm | color-convert | 2.0.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/color-convert/package.json |
| npm | color-name | 1.1.4 | MIT | /usr/local/lib/node_modules/npm/node_modules/color-name/package.json |
| npm | common-ancestor-path | 1.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/common-ancestor-path/package.json |
| npm | corepack | 0.36.0 | MIT | /usr/local/lib/node_modules/corepack/package.json |
| npm | cross-spawn | 7.0.6 | MIT | /usr/local/lib/node_modules/npm/node_modules/cross-spawn/package.json |
| npm | cssesc | 3.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/cssesc/package.json |
| npm | debug | 4.4.3 | MIT | /usr/local/lib/node_modules/npm/node_modules/debug/package.json |
| npm | diff | 5.2.2 | BSD-3-Clause | /usr/local/lib/node_modules/npm/node_modules/diff/package.json |
| npm | eastasianwidth | 0.2.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/eastasianwidth/package.json |
| npm | emoji-regex | 8.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/emoji-regex/package.json |
| npm | emoji-regex | 9.2.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/@isaacs/cliui/node_modules/emoji-regex/package.json; /usr/local/lib/node_modules/npm/node_modules/wrap-ansi/node_modules/emoji-regex/package.json |
| npm | encoding | 0.1.13 | MIT | /usr/local/lib/node_modules/npm/node_modules/encoding/package.json |
| npm | env-paths | 2.2.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/env-paths/package.json |
| npm | err-code | 2.0.3 | MIT | /usr/local/lib/node_modules/npm/node_modules/err-code/package.json |
| npm | exponential-backoff | 3.1.3 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/exponential-backoff/package.json |
| npm | fastest-levenshtein | 1.0.16 | MIT | /usr/local/lib/node_modules/npm/node_modules/fastest-levenshtein/package.json |
| npm | fdir | 6.5.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/fdir/package.json |
| npm | foreground-child | 3.3.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/foreground-child/package.json |
| npm | fs-minipass | 3.0.3 | ISC | /usr/local/lib/node_modules/npm/node_modules/fs-minipass/package.json |
| npm | glob | 10.5.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/glob/package.json |
| npm | graceful-fs | 4.2.11 | ISC | /usr/local/lib/node_modules/npm/node_modules/graceful-fs/package.json |
| npm | hosted-git-info | 8.1.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/hosted-git-info/package.json |
| npm | http-cache-semantics | 4.2.0 | BSD-2-Clause | /usr/local/lib/node_modules/npm/node_modules/http-cache-semantics/package.json |
| npm | http-proxy-agent | 7.0.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/http-proxy-agent/package.json |
| npm | https-proxy-agent | 7.0.6 | MIT | /usr/local/lib/node_modules/npm/node_modules/https-proxy-agent/package.json |
| npm | iconv-lite | 0.6.3 | MIT | /usr/local/lib/node_modules/npm/node_modules/iconv-lite/package.json |
| npm | ignore-walk | 7.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/ignore-walk/package.json |
| npm | imurmurhash | 0.1.4 | MIT | /usr/local/lib/node_modules/npm/node_modules/imurmurhash/package.json |
| npm | ini | 5.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/ini/package.json |
| npm | init-package-json | 7.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/init-package-json/package.json |
| npm | ip-address | 10.1.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/ip-address/package.json |
| npm | ip-regex | 5.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/ip-regex/package.json |
| npm | is-cidr | 5.1.1 | BSD-2-Clause | /usr/local/lib/node_modules/npm/node_modules/is-cidr/package.json |
| npm | is-fullwidth-code-point | 3.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/is-fullwidth-code-point/package.json |
| npm | isexe | 2.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/isexe/package.json |
| npm | isexe | 3.1.5 | BlueOak-1.0.0 | /usr/local/lib/node_modules/npm/node_modules/which/node_modules/isexe/package.json |
| npm | jackspeak | 3.4.3 | BlueOak-1.0.0 | /usr/local/lib/node_modules/npm/node_modules/jackspeak/package.json |
| npm | json-parse-even-better-errors | 4.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/json-parse-even-better-errors/package.json |
| npm | json-stringify-nice | 1.1.4 | ISC | /usr/local/lib/node_modules/npm/node_modules/json-stringify-nice/package.json |
| npm | jsonparse | 1.3.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/jsonparse/package.json |
| npm | just-diff | 6.0.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/just-diff/package.json |
| npm | just-diff-apply | 5.5.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/just-diff-apply/package.json |
| npm | libnpmaccess | 9.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmaccess/package.json |
| npm | libnpmdiff | 7.0.5 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmdiff/package.json |
| npm | libnpmexec | 9.0.5 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmexec/package.json |
| npm | libnpmfund | 6.0.5 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmfund/package.json |
| npm | libnpmhook | 11.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmhook/package.json |
| npm | libnpmorg | 7.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmorg/package.json |
| npm | libnpmpack | 8.0.5 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmpack/package.json |
| npm | libnpmpublish | 10.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmpublish/package.json |
| npm | libnpmsearch | 8.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmsearch/package.json |
| npm | libnpmteam | 7.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmteam/package.json |
| npm | libnpmversion | 7.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/libnpmversion/package.json |
| npm | lru-cache | 10.4.3 | ISC | /usr/local/lib/node_modules/npm/node_modules/lru-cache/package.json |
| npm | make-fetch-happen | 14.0.3 | ISC | /usr/local/lib/node_modules/npm/node_modules/make-fetch-happen/package.json |
| npm | minimatch | 9.0.9 | ISC | /usr/local/lib/node_modules/npm/node_modules/minimatch/package.json |
| npm | minipass | 3.3.6 | ISC | /usr/local/lib/node_modules/npm/node_modules/minipass-flush/node_modules/minipass/package.json; /usr/local/lib/node_modules/npm/node_modules/minipass-pipeline/node_modules/minipass/package.json; /usr/local/lib/node_modules/npm/node_modules/minipass-sized/node_modules/minipass/package.json |
| npm | minipass | 7.1.3 | BlueOak-1.0.0 | /usr/local/lib/node_modules/npm/node_modules/minipass/package.json |
| npm | minipass-collect | 2.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/minipass-collect/package.json |
| npm | minipass-fetch | 4.0.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/minipass-fetch/package.json |
| npm | minipass-flush | 1.0.5 | ISC | /usr/local/lib/node_modules/npm/node_modules/minipass-flush/package.json |
| npm | minipass-pipeline | 1.2.4 | ISC | /usr/local/lib/node_modules/npm/node_modules/minipass-pipeline/package.json |
| npm | minipass-sized | 1.0.3 | ISC | /usr/local/lib/node_modules/npm/node_modules/minipass-sized/package.json |
| npm | minizlib | 3.1.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/minizlib/package.json |
| npm | ms | 2.1.3 | MIT | /usr/local/lib/node_modules/npm/node_modules/ms/package.json |
| npm | mute-stream | 2.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/mute-stream/package.json |
| npm | negotiator | 1.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/negotiator/package.json |
| npm | node-gyp | 11.5.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/node-gyp/package.json |
| npm | nopt | 8.1.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/nopt/package.json |
| npm | normalize-package-data | 7.0.1 | BSD-2-Clause | /usr/local/lib/node_modules/npm/node_modules/normalize-package-data/package.json |
| npm | npm | 10.9.9 | Artistic-2.0 | /usr/local/lib/node_modules/npm/package.json |
| npm | npm-audit-report | 6.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/npm-audit-report/package.json |
| npm | npm-bundled | 4.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/npm-bundled/package.json |
| npm | npm-install-checks | 7.1.2 | BSD-2-Clause | /usr/local/lib/node_modules/npm/node_modules/npm-install-checks/package.json |
| npm | npm-normalize-package-bin | 4.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/npm-normalize-package-bin/package.json |
| npm | npm-package-arg | 12.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/npm-package-arg/package.json |
| npm | npm-packlist | 9.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/npm-packlist/package.json |
| npm | npm-pick-manifest | 10.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/npm-pick-manifest/package.json |
| npm | npm-profile | 11.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/npm-profile/package.json |
| npm | npm-registry-fetch | 18.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/npm-registry-fetch/package.json |
| npm | npm-user-validate | 3.0.0 | BSD-2-Clause | /usr/local/lib/node_modules/npm/node_modules/npm-user-validate/package.json |
| npm | p-map | 7.0.4 | MIT | /usr/local/lib/node_modules/npm/node_modules/p-map/package.json |
| npm | package-json-from-dist | 1.0.1 | BlueOak-1.0.0 | /usr/local/lib/node_modules/npm/node_modules/package-json-from-dist/package.json |
| npm | pacote | 19.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/pacote/package.json |
| npm | pacote | 20.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/@npmcli/metavuln-calculator/node_modules/pacote/package.json |
| npm | parse-conflict-json | 4.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/parse-conflict-json/package.json |
| npm | path-key | 3.1.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/path-key/package.json |
| npm | path-scurry | 1.11.1 | BlueOak-1.0.0 | /usr/local/lib/node_modules/npm/node_modules/path-scurry/package.json |
| npm | picomatch | 4.0.3 | MIT | /usr/local/lib/node_modules/npm/node_modules/picomatch/package.json |
| npm | postcss-selector-parser | 7.1.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/postcss-selector-parser/package.json |
| npm | proc-log | 5.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/proc-log/package.json |
| npm | proggy | 3.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/proggy/package.json |
| npm | promise-all-reject-late | 1.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/promise-all-reject-late/package.json |
| npm | promise-call-limit | 3.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/promise-call-limit/package.json |
| npm | promise-retry | 2.0.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/promise-retry/package.json |
| npm | promzard | 2.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/promzard/package.json |
| npm | qrcode-terminal | 0.12.0 | Apache 2.0 | /usr/local/lib/node_modules/npm/node_modules/qrcode-terminal/package.json |
| npm | read | 4.1.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/read/package.json |
| npm | read-cmd-shim | 5.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/read-cmd-shim/package.json |
| npm | read-package-json-fast | 4.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/read-package-json-fast/package.json |
| npm | retry | 0.12.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/retry/package.json |
| npm | safer-buffer | 2.1.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/safer-buffer/package.json |
| npm | semver | 7.7.4 | ISC | /usr/local/lib/node_modules/npm/node_modules/semver/package.json |
| npm | shebang-command | 2.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/shebang-command/package.json |
| npm | shebang-regex | 3.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/shebang-regex/package.json |
| npm | signal-exit | 4.1.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/signal-exit/package.json |
| npm | sigstore | 3.1.0 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/sigstore/package.json |
| npm | smart-buffer | 4.2.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/smart-buffer/package.json |
| npm | socks | 2.8.7 | MIT | /usr/local/lib/node_modules/npm/node_modules/socks/package.json |
| npm | socks-proxy-agent | 8.0.5 | MIT | /usr/local/lib/node_modules/npm/node_modules/socks-proxy-agent/package.json |
| npm | spdx-correct | 3.2.0 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/spdx-correct/package.json |
| npm | spdx-exceptions | 2.5.0 | CC-BY-3.0 | /usr/local/lib/node_modules/npm/node_modules/spdx-exceptions/package.json |
| npm | spdx-expression-parse | 3.0.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/spdx-correct/node_modules/spdx-expression-parse/package.json; /usr/local/lib/node_modules/npm/node_modules/validate-npm-package-license/node_modules/spdx-expression-parse/package.json |
| npm | spdx-expression-parse | 4.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/spdx-expression-parse/package.json |
| npm | spdx-license-ids | 3.0.23 | CC0-1.0 | /usr/local/lib/node_modules/npm/node_modules/spdx-license-ids/package.json |
| npm | ssri | 12.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/ssri/package.json |
| npm | string-width | 4.2.3 | MIT | /usr/local/lib/node_modules/npm/node_modules/string-width-cjs/package.json; /usr/local/lib/node_modules/npm/node_modules/string-width/package.json |
| npm | string-width | 5.1.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/@isaacs/cliui/node_modules/string-width/package.json; /usr/local/lib/node_modules/npm/node_modules/wrap-ansi/node_modules/string-width/package.json |
| npm | strip-ansi | 6.0.1 | MIT | /usr/local/lib/node_modules/npm/node_modules/strip-ansi-cjs/package.json; /usr/local/lib/node_modules/npm/node_modules/strip-ansi/package.json |
| npm | strip-ansi | 7.2.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/@isaacs/cliui/node_modules/strip-ansi/package.json; /usr/local/lib/node_modules/npm/node_modules/wrap-ansi/node_modules/strip-ansi/package.json |
| npm | supports-color | 9.4.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/supports-color/package.json |
| npm | tar | 7.5.22 | BlueOak-1.0.0 | /usr/local/lib/node_modules/npm/node_modules/tar/package.json |
| npm | text-table | 0.2.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/text-table/package.json |
| npm | tiny-relative-date | 1.3.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/tiny-relative-date/package.json |
| npm | tinyglobby | 0.2.15 | MIT | /usr/local/lib/node_modules/npm/node_modules/tinyglobby/package.json |
| npm | treeverse | 3.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/treeverse/package.json |
| npm | tuf-js | 3.1.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/tuf-js/package.json |
| npm | unique-filename | 4.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/unique-filename/package.json |
| npm | unique-slug | 5.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/unique-slug/package.json |
| npm | util-deprecate | 1.0.2 | MIT | /usr/local/lib/node_modules/npm/node_modules/util-deprecate/package.json |
| npm | validate-npm-package-license | 3.0.4 | Apache-2.0 | /usr/local/lib/node_modules/npm/node_modules/validate-npm-package-license/package.json |
| npm | validate-npm-package-name | 6.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/validate-npm-package-name/package.json |
| npm | walk-up-path | 3.0.1 | ISC | /usr/local/lib/node_modules/npm/node_modules/walk-up-path/package.json |
| npm | which | 2.0.2 | ISC | /usr/local/lib/node_modules/npm/node_modules/cross-spawn/node_modules/which/package.json |
| npm | which | 5.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/which/package.json |
| npm | wrap-ansi | 7.0.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/wrap-ansi-cjs/package.json |
| npm | wrap-ansi | 8.1.0 | MIT | /usr/local/lib/node_modules/npm/node_modules/wrap-ansi/package.json |
| npm | write-file-atomic | 6.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/write-file-atomic/package.json |
| npm | yallist | 4.0.0 | ISC | /usr/local/lib/node_modules/npm/node_modules/minipass-flush/node_modules/yallist/package.json; /usr/local/lib/node_modules/npm/node_modules/minipass-pipeline/node_modules/yallist/package.json; /usr/local/lib/node_modules/npm/node_modules/minipass-sized/node_modules/yallist/package.json |
| npm | yallist | 5.0.0 | BlueOak-1.0.0 | /usr/local/lib/node_modules/npm/node_modules/yallist/package.json |
| npm | yarn | 1.22.22 | BSD-2-Clause | /opt/yarn-v1.22.22/package.json |

</details>

## Supplementary Codex native-library inventory

Codex ships native voice libraries even though ssh-ballpit does not use voice. These are easy to miss in package-manager scans. The following versions and source hashes come from the image's `codex-resources/voice/sources.json`; license labels were checked against its accompanying `NOTICE.md` and `licenses/` files. LGPL version labels identify the supplied license texts; consult upstream files for any later-version options.

Meson and Ninja entries in that manifest are upstream build inputs, not bundled runtime libraries, and are excluded from this supplementary runtime table. GVDB is included with GLib under LGPL; sljit is included with PCRE2 under BSD-2-Clause.

| Component | Version / revision | License evidence | Source archive SHA256 |
| --- | --- | --- | --- |
| [libffi](https://github.com/libffi/libffi/releases/download/v3.8.0/libffi-3.8.0.tar.gz) | 3.8.0 | MIT; licenses/libffi.txt | 7da3e2d9a171eb0a038f592ecad3ff2bb2550f3496d87b3b29ad0cf4430c0db4 |
| [pcre2](https://github.com/PCRE2Project/pcre2/releases/download/pcre2-10.47/pcre2-10.47.tar.gz) | 10.47 | BSD-3-Clause WITH PCRE2-exception; bundled sljit: BSD-2-Clause; licenses/PCRE2.md and sljit.txt | c08ae2388ef333e8403e670ad70c0a11f1eed021fd88308d7e02f596fcd9dc16 |
| [zlib](https://github.com/madler/zlib/releases/download/v1.3.2/zlib-1.3.2.tar.gz) | 1.3.2 | Zlib; licenses/zlib.txt | bb329a0a2cd0274d05519d61c667c062e06990d72e125ee2dfa8de64f0119d16 |
| [opus](https://downloads.xiph.org/releases/opus/opus-1.6.1.tar.gz) | 1.6.1 | BSD-3-Clause; licenses/Opus.txt | 6ffcb593207be92584df15b32466ed64bbec99109f007c82205f0194572411a1 |
| [glib](https://download.gnome.org/sources/glib/2.88/glib-2.88.3.tar.xz) | 2.88.3 | LGPL-2.1; licenses/LGPL-2.1.txt and NOTICE.md | ab24d24e698dfa1e408b7bcdb508f4aafc906185a8b8ce72fdf79bbbdc9b383b |
| [proxy-libintl](https://github.com/frida/proxy-libintl/archive/refs/tags/0.5.tar.gz) | 0.5 | LGPL-2.0; licenses/proxy-libintl.txt | f7a1cbd7579baaf575c66f9d99fb6295e9b0684a28b095967cfda17857595303 |
| [gstreamer](https://gstreamer.freedesktop.org/src/gstreamer/gstreamer-1.28.6.tar.xz) | 1.28.6 | LGPL-2.1; licenses/LGPL-2.1.txt and NOTICE.md | 62b6b9f0ad3147a6dd6420ac64a91180b14e990695bddd353b96041611d052ca |
| [gst-plugins-base](https://gstreamer.freedesktop.org/src/gst-plugins-base/gst-plugins-base-1.28.6.tar.xz) | 1.28.6 | LGPL-2.1; licenses/LGPL-2.1.txt and NOTICE.md | 0ba699c7c6c66f4ba640be78cb38a24715add9683f3e3a199f5369dc5a4f04ac |
| [gst-plugins-good](https://gstreamer.freedesktop.org/src/gst-plugins-good/gst-plugins-good-1.28.6.tar.xz) | 1.28.6 | LGPL-2.1; licenses/LGPL-2.1.txt and NOTICE.md | b0c620a4b18b6ee931b4c43bbf1760d308666dc37f730a7e7f1ad327e59ce2df |
| gvdb | 2b42fc75f09dbe1cd1057580b5782b08f2dcb400 | LGPL; included by glib | Included in parent archive |

Image evidence: `usr/local/lib/node_modules/@openai/codex/node_modules/@openai/codex-linux-x64/vendor/x86_64-unknown-linux-musl/codex-resources/voice/sources.json` and the adjacent `NOTICE.md`, `manifest.json`, and `licenses/`.

## Refresh

Install Go, Python 3, rootless Podman, and [Syft](https://github.com/anchore/syft). Run from the repository root; the archive can be stored outside the checkout:

```sh
podman build -t localhost/ssh-ballpit:sbom .
podman save --format oci-archive -o /tmp/ssh-ballpit-sbom.oci localhost/ssh-ballpit:sbom
python3 scripts/generate-sbom.py --image-archive /tmp/ssh-ballpit-sbom.oci
```

`GO` and `SYFT` can select explicit tool paths. Generation downloads the selected module graph, may add missing checksums to `go.sum`, and uses network access for license enrichment. Review changed licenses and any `NOASSERTION` entries, and update the README license summary when dependencies or the container recipe change.
