post_upgrade() {
    systemctl daemon-reload
}

post_install() {
    systemctl daemon-reload
}
