(function () {
  window.dashboardPage = function () {
    return {
      tab: 'vin',
      vin: '',
      plate: '',
      state: '',
      submitting: false,
      purchasing: false,
      error: '',
      noCreditsOpen: false,
      credits: 0,
      lookupURL: '',
      purchaseURL: '',
      confirm: {
        open: false,
        token: '',
        vehicle: '',
        vin: '',
        records: 0,
      },

      init() {
        this.credits = parseInt(this.$el.dataset.credits || '0', 10);
        this.lookupURL = this.$el.dataset.lookupUrl || '';
        this.purchaseURL = this.$el.dataset.purchaseUrl || '';
      },

      csrfToken() {
        const meta = document.querySelector('meta[name="csrf-token"]');
        return meta ? meta.getAttribute('content') : '';
      },

      async post(url, formData) {
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            Accept: 'application/json',
            'X-CSRF-Token': this.csrfToken(),
          },
          body: formData,
        });
        const data = await response.json().catch(function () {
          return {};
        });
        if (!response.ok) {
          throw new Error(data.error || 'Something went wrong. Please try again.');
        }
        return data;
      },

      async submitLookup() {
        if (this.credits < 1) {
          this.noCreditsOpen = true;
          return;
        }

        this.submitting = true;
        this.error = '';

        const formData = new FormData();
        if (this.tab === 'vin') {
          formData.set('vin', this.vin);
        } else {
          formData.set('plate', this.plate);
          formData.set('state', this.state);
        }

        try {
          const data = await this.post(this.lookupURL, formData);
          this.confirm = {
            open: true,
            token: data.token,
            vehicle: data.vehicle || 'Vehicle report',
            vin: data.vin,
            records: data.records,
          };
        } catch (err) {
          this.error = err.message;
        } finally {
          this.submitting = false;
        }
      },

      async purchase() {
        this.purchasing = true;
        this.error = '';

        const formData = new FormData();
        formData.set('token', this.confirm.token);

        try {
          const data = await this.post(this.purchaseURL, formData);
          if (!data.redirect) {
            throw new Error('Could not complete the purchase. Please try again.');
          }
          window.location.assign(data.redirect);
        } catch (err) {
          this.error = err.message;
          this.purchasing = false;
        }
      },
    };
  };
})();
