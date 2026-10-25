(function () {
  const HISTORY_KEY = 'carf_chorf_recent_vins';
  const MAX_RECENT_VINS = 5;
  const VIN_PATTERN = /^[A-HJ-NPR-Z0-9]{17}$/;

  function isValidVinFormat(value) {
    return VIN_PATTERN.test(value);
  }

  function loadHistory() {
    try {
      const saved = JSON.parse(localStorage.getItem(HISTORY_KEY) || '[]');
      if (!Array.isArray(saved)) {
        return [];
      }

      return saved
        .filter((vin) => typeof vin === 'string')
        .map((vin) => vin.trim().toUpperCase())
        .filter((vin, index, list) => isValidVinFormat(vin) && list.indexOf(vin) === index)
        .slice(0, MAX_RECENT_VINS);
    } catch (error) {
      return [];
    }
  }

  function saveHistory(vins) {
    try {
      localStorage.setItem(HISTORY_KEY, JSON.stringify(vins));
    } catch (error) {
      // localStorage can be unavailable in private browsing; keep in-memory history.
    }
  }

  window.vehicleLookup = function () {
    return {
      tab: 'vin',
      submitting: false,
      vin: '',
      plate: '',
      recentVins: [],
      vinHistoryOpen: false,

      init() {
        this.tab = this.$el.dataset.tab || 'vin';
        this.vin = this.$el.dataset.vin || '';
        this.plate = this.$el.dataset.plate || '';
        this.recentVins = loadHistory();
      },

      get filteredVins() {
        const query = this.vin.trim().toUpperCase();
        if (!query) {
          return this.recentVins;
        }
        return this.recentVins.filter((vin) => vin.includes(query));
      },

      chooseVin(value) {
        this.vin = value;
        this.tab = 'vin';
        this.vinHistoryOpen = false;
      },

      clearVins() {
        this.recentVins = [];
        this.vinHistoryOpen = false;
        try {
          localStorage.removeItem(HISTORY_KEY);
        } catch (error) {
          // Ignore storage errors.
        }
      },

      rememberVin() {
        if (this.tab !== 'vin') {
          return;
        }

        const value = this.vin.trim().toUpperCase();
        if (!isValidVinFormat(value)) {
          return;
        }

        this.recentVins = [value, ...this.recentVins.filter((vin) => vin !== value)].slice(0, MAX_RECENT_VINS);
        saveHistory(this.recentVins);
      },
    };
  };
})();
