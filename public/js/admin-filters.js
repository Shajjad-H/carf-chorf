(function () {
  var debounceDelay = 400;

  document.querySelectorAll('form[data-auto-filter]').forEach(function (form) {
    var timer = null;
    var submitting = false;

    function submit() {
      if (submitting) {
        return;
      }
      submitting = true;
      if (typeof form.requestSubmit === 'function') {
        form.requestSubmit();
      } else {
        form.submit();
      }
    }

    form.addEventListener('submit', function () {
      clearTimeout(timer);
      submitting = true;
    });

    form.querySelectorAll('input, select').forEach(function (field) {
      if (field.type === 'button' || field.type === 'submit' || field.name === 'csrf_token') {
        return;
      }

      if (field.tagName === 'SELECT' || field.type === 'date' || field.type === 'number') {
        field.addEventListener('change', function () {
          clearTimeout(timer);
          submit();
        });
        return;
      }

      field.addEventListener('input', function () {
        clearTimeout(timer);
        timer = setTimeout(submit, debounceDelay);
      });
    });
  });
})();
