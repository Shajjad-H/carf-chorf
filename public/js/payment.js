(function () {
  window.paymentPage = function () {
    return {
      token: '',
      createOrderURL: '',
      captureOrderURL: '',
      sdkReady: false,
      processing: false,
      showCardFields: true,
      errors: [],
      cardFields: null,

      init() {
        this.token = this.$el.dataset.token || '';
        this.createOrderURL = this.$el.dataset.createOrderUrl || '';
        this.captureOrderURL = this.$el.dataset.captureOrderUrl || '';

        if (typeof window.paypal === 'undefined') {
          this.showCardFields = false;
          this.errors.push('Payment is temporarily unavailable. Please try again later.');
          return;
        }

        this.sdkReady = true;
        this.renderCardFields();
        this.renderPayPalButtons();
      },

      csrfToken() {
        const meta = document.querySelector('meta[name="csrf-token"]');
        return meta ? meta.getAttribute('content') : '';
      },

      async post(url) {
        const response = await fetch(url + '?token=' + encodeURIComponent(this.token), {
          method: 'POST',
          headers: {
            Accept: 'application/json',
            'X-CSRF-Token': this.csrfToken(),
          },
        });
        const data = await response.json().catch(function () {
          return {};
        });
        if (!response.ok) {
          throw new Error(data.error || 'Payment could not be completed. Please try again.');
        }
        return data;
      },

      async createOrder() {
        const data = await this.post(this.createOrderURL);
        if (!data.pid) {
          throw new Error('Payment could not be started. Please try again.');
        }
        return data.pid;
      },

      async captureOrder() {
        this.processing = true;
        try {
          const data = await this.post(this.captureOrderURL);
          if (!data.redirect) {
            throw new Error('Payment could not be completed. Please try again.');
          }
          window.location.assign(data.redirect);
        } catch (error) {
          this.processing = false;
          this.handleError(error);
          throw error;
        }
      },

      async submitCard() {
        if (this.processing || !this.cardFields) {
          return;
        }

        this.processing = true;
        this.errors = [];
        try {
          const result = await this.cardFields.submit();
          if (result && result.error) {
            throw new Error(result.error.message || 'Card payment failed.');
          }
        } catch (error) {
          this.processing = false;
          this.handleError(error);
        }
      },

      handleError(error) {
        let message = error && error.message ? error.message : 'Payment could not be completed. Please try again.';

        if (message === 'INVALID_NUMBER') {
          message = 'Invalid card number. Please check the details and try again.';
        } else if (message === 'DECLINED') {
          message = 'The payment was declined. Please try another payment method.';
        }

        if (message.length > 180) {
          message = 'Payment could not be completed. Please try again.';
        }
        if (!this.errors.includes(message)) {
          this.errors.push(message);
        }
      },

      renderCardFields() {
        if (typeof window.paypal.CardFields !== 'function') {
          this.showCardFields = false;
          return;
        }

        try {
          this.cardFields = window.paypal.CardFields({
            createOrder: () => this.createOrder(),
            onApprove: () => this.captureOrder(),
            onError: (error) => this.handleError(error),
            style: {
              input: {
                'padding': '0.75rem 1rem',
                'font-size': '0.9375rem',
                'border': '2px solid #e2e8f0',
                'border-radius': '0.5rem',
                'background': 'white',
              },
            },
          });

          if (!this.cardFields.isEligible()) {
            this.showCardFields = false;
            return;
          }

          this.cardFields.NameField({ placeholder: 'John Doe' }).render('#card-name-field-container');
          this.cardFields.NumberField({ placeholder: '1234 5678 9012 3456' }).render('#card-number-field-container');
          this.cardFields.CVVField().render('#card-cvv-field-container');
          this.cardFields.ExpiryField().render('#card-expiry-field-container');
        } catch (error) {
          this.showCardFields = false;
        }
      },

      renderPayPalButtons() {
        if (typeof window.paypal.Buttons !== 'function') {
          return;
        }

        try {
          const button = window.paypal.Buttons({
            createOrder: () => this.createOrder(),
            onApprove: () => this.captureOrder(),
            onError: (error) => this.handleError(error),
            onCancel: () => {
              this.processing = false;
            },
          });

          if (button.isEligible()) {
            button.render('#paypal-button-container');
          }
        } catch (error) {
          this.handleError(error);
        }
      },
    };
  };
})();
