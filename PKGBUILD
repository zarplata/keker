# Maintainer: Andrey Platonov <a.platonov@office.ngs.ru>
pkgname=keker
_binaryname=kekerd
pkgver=${VERSION:-manual}
pkgrel=1
pkgdesc="Simple pub-sub message broker via websockets"
arch=("any")
license=('MIT')
groups=()
makedepends=('go')
provides=()
replaces=()
branch=${BRANCH:-master}
install='install.sh'
options=(emptydirs)
source=("git+https://github.com/tears-of-noobs/keker.git#branch=$branch")
md5sums=('SKIP')


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
