# Maintainer: Andrey Platonov <a.platonov@office.ngs.ru>
pkgname=keker
_binaryname=kekerd
pkgver=20170411.30_c30c722
pkgrel=1
pkgdesc="Simple pub-sub message broker via websockets"
arch=("any")
license=('MIT')
groups=()
makedepends=('go')
provides=()
replaces=()
install='install.sh'
options=(emptydirs)
source=("git+https://github.com/tears-of-noobs/keker.git#branch=master")
md5sums=('SKIP')

pkgver() {
    cd "$pkgname"
    make ver
}

build() {
	cd "$srcdir/$pkgname"
    make
}
package() {
    PKG_SYSTEMD="$pkgdir/usr/lib/systemd/system"
    cd "$srcdir/$pkgname"

    install -Dm 0755 ".out/$_binaryname" "${pkgdir}/usr/bin/$_binaryname"
    install -Dm 0644 "../../$_binaryname.service" "$PKG_SYSTEMD/$_binaryname.service"
}
