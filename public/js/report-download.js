(function () {
  var pdfPromise = null;

  function reportIsReady() {
    if (!document.body) {
      return false;
    }
    if ((document.body.innerText || '').trim().length > 800) {
      return true;
    }
    var root = document.getElementById('body-element');
    return !!(root && root.querySelector('main, [class*="report"], [id*="Section"]'));
  }

  function waitForReport(timeoutMs) {
    return new Promise(function (resolve, reject) {
      var started = Date.now();
      (function check() {
        if (reportIsReady()) {
          resolve();
          return;
        }
        if (Date.now() - started > timeoutMs) {
          reject(new Error('Report is not ready'));
          return;
        }
        setTimeout(check, 200);
      })();
    });
  }

  function expandSections() {
    document.querySelectorAll('details:not([open])').forEach(function (el) {
      el.open = true;
    });
    document.querySelectorAll('button[aria-expanded="false"]').forEach(function (el) {
      try {
        el.click();
      } catch (e) {}
    });
  }

  function prepareLogoImages() {
    var logos = document.querySelectorAll('img[src*="CarfaxLogo.svg"]');
    if (!logos.length) {
      return Promise.resolve();
    }
    var src = logos[0].src;
    return fetch(src).then(function (response) {
      if (!response.ok) {
        throw new Error('logo fetch failed');
      }
      return response.text();
    }).then(function (svg) {
      return new Promise(function (resolve, reject) {
        var blob = new Blob([svg], { type: 'image/svg+xml' });
        var url = URL.createObjectURL(blob);
        var image = new Image();
        image.onload = function () {
          try {
            var ratio = 2;
            var width = 378.51;
            var height = 70.97;
            var canvas = document.createElement('canvas');
            canvas.width = Math.round(width * ratio);
            canvas.height = Math.round(height * ratio);
            var ctx = canvas.getContext('2d');
            ctx.scale(ratio, ratio);
            ctx.drawImage(image, 0, 0, width, height);
            var data = canvas.toDataURL('image/png');
            logos.forEach(function (logo) {
              logo.src = data;
            });
            URL.revokeObjectURL(url);
            resolve();
          } catch (error) {
            URL.revokeObjectURL(url);
            reject(error);
          }
        };
        image.onerror = function () {
          URL.revokeObjectURL(url);
          reject(new Error('logo load failed'));
        };
        image.src = url;
      });
    });
  }

  function loadHtml2Pdf() {
    if (window.html2pdf) {
      return Promise.resolve(window.html2pdf);
    }
    return new Promise(function (resolve, reject) {
      var script = document.createElement('script');
      script.src = '/public/vendor/html2pdf.bundle.min.js';
      script.async = true;
      script.onload = function () {
        if (window.html2pdf) {
          resolve(window.html2pdf);
        } else {
          reject(new Error('html2pdf unavailable'));
        }
      };
      script.onerror = function () {
        reject(new Error('html2pdf failed to load'));
      };
      document.head.appendChild(script);
    });
  }

  function reportVin() {
    try {
      var data = window.__INITIAL__DATA__;
      var info = data && data.vhr && data.vhr.headerSection && data.vhr.headerSection.vehicleInformation;
      var vin = info && info.vin && (info.vin.vin || info.vin);
      return typeof vin === 'string' ? vin : '';
    } catch (e) {
      return '';
    }
  }

  function reportFilename() {
    var vin = reportVin().replace(/[^A-Za-z0-9_-]+/g, '');
    return 'vehicle-report-' + (vin || 'report') + '.pdf';
  }

  function canvasScale() {
    var height = Math.max(document.body.scrollHeight, document.documentElement.scrollHeight, 1);
    return Math.min(2, Math.max(0.75, 30000 / height));
  }

  function reportRoot() {
    return document.querySelector('.vhr-tabs') || document.body;
  }

  function canvasWidth() {
    var root = reportRoot();
    return Math.max(root.scrollWidth || 0, Math.round(root.getBoundingClientRect().width || 0), 1);
  }

  function cropCanvas(worker) {
    var canvas = worker && worker.prop && worker.prop.canvas;
    if (!canvas) {
      return;
    }
    var ctx = canvas.getContext('2d');
    var width = canvas.width;
    var height = canvas.height;
    var maxX = 0;
    var step = Math.max(1, Math.floor(height / 160));
    for (var y = 0; y < height; y += step) {
      var row;
      try {
        row = ctx.getImageData(0, y, width, 1).data;
      } catch (e) {
        return;
      }
      for (var x = width - 1; x >= 0; x -= 1) {
        var i = x * 4;
        if (row[i] < 250 || row[i + 1] < 250 || row[i + 2] < 250) {
          if (x > maxX) {
            maxX = x;
          }
          break;
        }
      }
    }
    var cropWidth = Math.min(width, maxX + 8);
    window.__pdfCropWidth = cropWidth;
    if (cropWidth > 0 && cropWidth < width) {
      var cropped = document.createElement('canvas');
      cropped.width = cropWidth;
      cropped.height = height;
      cropped.getContext('2d').drawImage(canvas, 0, 0);
      worker.prop.canvas = cropped;
    }
  }

  function generate() {
    return waitForReport(20000)
      .then(function () {
        expandSections();
        window.scrollTo(0, 0);
        return prepareLogoImages().catch(function () {}).then(loadHtml2Pdf);
      })
      .then(function (html2pdf) {
        var root = reportRoot();
        var width = canvasWidth();
        var worker = html2pdf().set({
          margin: [8, 8, 8, 8],
          filename: reportFilename(),
          image: { type: 'jpeg', quality: 0.95 },
          html2canvas: {
            scale: canvasScale(),
            useCORS: true,
            allowTaint: false,
            backgroundColor: '#ffffff',
            scrollX: 0,
            scrollY: 0,
            windowWidth: document.documentElement.scrollWidth || width,
            width: width
          },
          jsPDF: { unit: 'pt', format: 'letter', orientation: 'portrait' },
          pagebreak: { mode: ['css', 'legacy'] }
        }).from(root);
        return worker.toCanvas().then(function () {
          cropCanvas(worker);
          return worker.save();
        });
      });
  }

  window.downloadReportPDF = function () {
    if (!pdfPromise) {
      pdfPromise = generate().catch(function (error) {
        pdfPromise = null;
        throw error;
      });
    }
    return pdfPromise;
  };
})();
