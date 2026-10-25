(function () {
  window.pricingPage = function () {
    return {
      loginOpen: false,

      showLogin() {
        this.loginOpen = true;
      },

      goLogin() {
        this.goToAuth('/login');
      },

      goGetStarted() {
        this.goToAuth('/getstarted');
      },

      goToAuth(path) {
        const next = '/checkout';
        window.location.href = path + '?next=' + encodeURIComponent(next);
      },
    };
  };
})();
