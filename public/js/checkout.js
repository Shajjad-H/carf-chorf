(function () {
  window.checkoutPage = function () {
    return {
      selected: 0,
      authenticated: false,
      hasVehicle: false,
      termsAccepted: false,
      submitting: false,
      token: '',

      init() {
        const data = this.$el.dataset;
        this.selected = Number(data.selected || 0);
        this.authenticated = data.authenticated === 'true';
        this.hasVehicle = data.hasVehicle === 'true';
        this.token = data.token || '';

        window.addEventListener('pageshow', (event) => {
          if (event.persisted) {
            this.submitting = false;
          }
        });
      },

      selectPackage(credits) {
        if (!this.authenticated && !this.hasVehicle) {
          window.location.href = '/#lookup';
          return;
        }

        this.selected = credits;
      },
    };
  };
})();

// Champs email du checkout invité : suggestion de domaine mal tapé + vérification de la confirmation.
// Purement visuel : la validation serveur (ErrInvalidEmail, ErrEmailMismatch) reste la référence.
(function () {
  var DOMAINS = ['gmail.com', 'yahoo.com', 'hotmail.com', 'outlook.com', 'icloud.com', 'aol.com', 'live.com', 'msn.com', 'comcast.net', 'proton.me', 'protonmail.com'];

  function distance(a, b) {
    if (Math.abs(a.length - b.length) > 2) return 3;
    var prev = [];
    for (var j = 0; j <= b.length; j++) prev.push(j);
    for (var i = 1; i <= a.length; i++) {
      var cur = [i];
      for (var k = 1; k <= b.length; k++) {
        cur[k] = Math.min(prev[k] + 1, cur[k - 1] + 1, prev[k - 1] + (a[i - 1] === b[k - 1] ? 0 : 1));
      }
      prev = cur;
    }
    return prev[b.length];
  }

  window.emailFields = function () {
    return {
      email: '',
      confirm: '',
      confirmTouched: false,

      get suggestion() {
        var at = this.email.lastIndexOf('@');
        if (at < 1) return '';
        var domain = this.email.slice(at + 1).toLowerCase();
        if (domain.length < 4 || DOMAINS.indexOf(domain) !== -1) return '';
        for (var i = 0; i < DOMAINS.length; i++) {
          var d = distance(domain, DOMAINS[i]);
          if (d > 0 && d <= 2) return this.email.slice(0, at + 1) + DOMAINS[i];
        }
        return '';
      },

      get matches() {
        return this.confirm !== '' && this.confirm.toLowerCase() === this.email.toLowerCase();
      },

      get showMismatch() {
        if (this.confirm === '' || this.matches) return false;
        return this.confirmTouched || this.confirm.length >= this.email.length;
      },

      applySuggestion() {
        var fixed = this.suggestion;
        if (this.confirm.toLowerCase() === this.email.toLowerCase()) this.confirm = fixed;
        this.email = fixed;
      },
    };
  };
})();
