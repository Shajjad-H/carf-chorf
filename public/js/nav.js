(function () {
  window.navShell = function () {
    return {
      mobileOpen: false,
      scrolled: false,
      onScroll: null,

      init() {
        this.onScroll = () => {
          this.scrolled = window.scrollY > 8;
        };
        this.onScroll();
        window.addEventListener('scroll', this.onScroll, { passive: true });
      },

      destroy() {
        if (this.onScroll) {
          window.removeEventListener('scroll', this.onScroll);
        }
      },

      open() {
        this.mobileOpen = true;
        this.$nextTick(() => {
          if (this.$refs.closeButton) {
            this.$refs.closeButton.focus();
          }
        });
      },

      close() {
        this.mobileOpen = false;
      },
    };
  };
})();
