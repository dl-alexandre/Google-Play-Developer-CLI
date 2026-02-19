pkgname=gpd-git
pkgver=0.1.0.2.gd2f8b0d
pkgrel=1
pkgdesc="Google Play Developer CLI"
arch=('x86_64' 'aarch64')
url="https://github.com/dl-alexandre/Google-Play-Developer-CLI"
license=('Apache-2.0')
depends=('libsecret')
makedepends=('git' 'go')
provides=('gpd')
conflicts=('gpd')
source=("git+https://github.com/dl-alexandre/Google-Play-Developer-CLI.git")
sha256sums=('SKIP')

pkgver() {
	cd "$srcdir/gpd"
	git describe --tags --long --match 'v*' | sed 's/^v//;s/-/./g'
}

build() {
	cd "$srcdir/gpd"
	local commit
	commit="$(git rev-parse --short HEAD)"
	local ldflags="-s -w -buildid= -X github.com/dl-alexandre/gpd/pkg/version.Version=${pkgver} -X github.com/dl-alexandre/gpd/pkg/version.GitCommit=${commit} -X github.com/dl-alexandre/gpd/pkg/version.BuildTime=unknown"
	go build -trimpath -buildmode=pie -mod=readonly -ldflags "$ldflags" -o gpd ./cmd/gpd
}

package() {
	cd "$srcdir/gpd"
	install -Dm755 gpd "$pkgdir/usr/bin/gpd"
	install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
