(function () {
  const CHECKOUT_PATH = '/checkout';

  function checkoutURL(token, credits) {
    const params = new URLSearchParams();
    if (token) {
      params.set('token', token);
    }
    if (credits) {
      params.set('credits', credits);
    }
    return CHECKOUT_PATH + (params.toString() ? '?' + params.toString() : '');
  }

  window.checkoutPage = function () {
    return {
      selected: 0,
      pendingCredits: 0,
      loginOpen: false,
      authenticated: false,
      hasVehicle: false,
      termsAccepted: false,
      submitting: false,
      token: '',

      init() {
        const data = this.$el.dataset;
        this.selected = Number(data.selected || 0);
        this.pendingCredits = Number(data.requested || 0);
        this.loginOpen = data.loginOpen === 'true';
        this.authenticated = data.authenticated === 'true';
        this.hasVehicle = data.hasVehicle === 'true';
        this.token = data.token || '';
      },

      selectPackage(credits) {
        if (!this.authenticated && credits > 1) {
          this.pendingCredits = credits;
          this.loginOpen = true;
          return;
        }

        if (!this.authenticated && credits === 1 && !this.hasVehicle) {
          window.location.href = '/';
          return;
        }

        this.selected = credits;
      },

      packageActionLabel(credits) {
        if (this.authenticated) {
          return this.selected === credits ? 'Selected' : 'Select package';
        }
        if (credits > 1) {
          return 'Login to choose';
        }
        if (!this.hasVehicle) {
          return 'Look up a vehicle';
        }
        return this.selected === credits ? 'Selected' : 'Buy 1 credit';
      },

      goLogin() {
        this.goToAuth('/login');
      },

      goGetStarted() {
        this.goToAuth('/getstarted');
      },

      goToAuth(path) {
        const next = checkoutURL(this.token, this.pendingCredits);
        window.location.href = path + '?next=' + encodeURIComponent(next);
      },
    };
  };
})();
