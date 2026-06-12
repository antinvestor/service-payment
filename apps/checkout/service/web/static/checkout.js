(function () {
  'use strict';

  // --- polling for confirm page ---
  var body = document.body;
  var pollURL = body && body.dataset.poll;
  var returnURL = body && body.dataset.return;

  if (pollURL) {
    var intervalID = setInterval(function () {
      fetch(pollURL)
        .then(function (res) { return res.json(); })
        .then(function (data) {
          if (data.status === 'completed') {
            clearInterval(intervalID);
            var dest = returnURL || location.href;
            var sep = dest.indexOf('?') >= 0 ? '&' : '?';
            location.href = dest + sep + 'status=completed';
          } else if (data.status === 'failed') {
            clearInterval(intervalID);
            location.reload();
          }
        })
        .catch(function () { /* network error — keep polling */ });
    }, 2000);
  }

  // --- change-contact toggle ---
  var changeBtn = document.getElementById('change');
  var contactEdit = document.getElementById('contact-edit');
  if (changeBtn && contactEdit) {
    changeBtn.addEventListener('click', function () {
      contactEdit.classList.toggle('hidden');
    });
  }

  // --- chip highlight sync (in case CSS :has() not supported) ---
  var chips = document.querySelectorAll('.chip');
  chips.forEach(function (chip) {
    var radio = chip.querySelector('input[type="radio"]');
    if (!radio) return;
    radio.addEventListener('change', function () {
      chips.forEach(function (c) { c.classList.remove('chip-selected'); });
      if (radio.checked) chip.classList.add('chip-selected');
    });
  });
}());
