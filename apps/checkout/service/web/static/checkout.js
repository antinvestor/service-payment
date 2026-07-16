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
          if (data.next_action === 'requires_pin') {
            showAuthStep('auth-pin');
            return;
          }
          if (data.next_action === 'requires_otp') {
            showAuthStep('auth-otp');
            return;
          }
          if (data.redirect_url) {
            // 3DS / external auth — try iframe first, else top-level (banks often block iframe).
            clearInterval(intervalID);
            if (!openAuthFrame(data.redirect_url)) {
              window.location = data.redirect_url;
            } else {
              // Keep a soft poll so we leave the frame when charge completes.
              intervalID = setInterval(function () {
                fetch(pollURL)
                  .then(function (r) { return r.json(); })
                  .then(function (d) {
                    if (d.status === 'completed' || d.status === 'failed') {
                      clearInterval(intervalID);
                      finish(d);
                    }
                  })
                  .catch(function () {});
              }, 2500);
            }
          } else if (data.status === 'completed' || data.status === 'failed') {
            clearInterval(intervalID);
            finish(data);
          }
        })
        .catch(function () { /* network error — keep polling */ });
    }, 2000);
  }

  function finish(data) {
    if (data.status === 'completed') {
      if (returnURL) {
        location.href = returnURL;
      } else {
        var dest = returnURL || location.href;
        var sep = dest.indexOf('?') >= 0 ? '&' : '?';
        var session = body.dataset.session || '';
        location.href = dest + sep + 'status=completed' +
          (session ? '&session=' + encodeURIComponent(session) : '');
      }
    } else {
      location.reload();
    }
  }

  function showAuthStep(id) {
    ['auth-pin', 'auth-otp'].forEach(function (sid) {
      var el = document.getElementById(sid);
      if (el) el.classList.toggle('hidden', sid !== id);
    });
    var spin = document.querySelector('.spinner');
    if (spin) spin.classList.add('hidden');
  }

  function openAuthFrame(url) {
    var wrap = document.getElementById('auth-frame-wrap');
    var frame = document.getElementById('auth-frame');
    if (!wrap || !frame) return false;
    try {
      frame.src = url;
      wrap.classList.remove('hidden');
      var spin = document.querySelector('.spinner');
      if (spin) spin.classList.add('hidden');
      return true;
    } catch (e) {
      return false;
    }
  }

  // --- change-contact toggle ---
  var changeBtn = document.getElementById('change');
  var contactEdit = document.getElementById('contact-edit');
  if (changeBtn && contactEdit) {
    changeBtn.addEventListener('click', function () {
      contactEdit.classList.toggle('hidden');
    });
  }

  // --- method chips: show/hide card fields ---
  var chips = document.querySelectorAll('.chip');
  var cardFields = document.getElementById('card-fields');
  var phoneField = document.getElementById('phone-field');
  var useSaved = document.getElementById('use-saved');

  function selectedChip() {
    for (var i = 0; i < chips.length; i++) {
      var radio = chips[i].querySelector('input[type="radio"]');
      if (radio && radio.checked) return chips[i];
    }
    return null;
  }

  function syncMethodUI() {
    var chip = selectedChip();
    var embed = chip && chip.getAttribute('data-embed') === '1';
    var key = chip && chip.getAttribute('data-key');
    chips.forEach(function (c) { c.classList.remove('chip-selected'); });
    if (chip) chip.classList.add('chip-selected');
    if (cardFields) {
      var showCard = embed && !(useSaved && useSaved.checked);
      cardFields.classList.toggle('hidden', !showCard);
    }
    if (phoneField && key) {
      // Phone only required for MoMo-style methods.
      phoneField.classList.toggle('hidden', embed || key === 'card');
    }
  }

  chips.forEach(function (chip) {
    var radio = chip.querySelector('input[type="radio"]');
    if (!radio) return;
    radio.addEventListener('change', syncMethodUI);
  });
  if (useSaved) {
    useSaved.addEventListener('change', syncMethodUI);
  }
  syncMethodUI();

  // --- card encryption on submit (AES-GCM, Flutterwave v4) ---
  var form = document.getElementById('pay-form');
  if (form) {
    form.addEventListener('submit', function (ev) {
      var chip = selectedChip();
      var embed = chip && chip.getAttribute('data-embed') === '1';
      if (!embed) return;
      if (useSaved && useSaved.checked) return; // token path — no PAN

      ev.preventDefault();
      var cryptoURL = form.getAttribute('data-crypto');
      if (!cryptoURL) {
        alert('Card encryption is not available. Please try again later.');
        return;
      }
      var pan = (document.getElementById('card-number') || {}).value || '';
      var exp = (document.getElementById('card-exp') || {}).value || '';
      var cvv = (document.getElementById('card-cvv') || {}).value || '';
      pan = pan.replace(/\s+/g, '');
      var parts = exp.split('/');
      var mm = (parts[0] || '').trim();
      var yy = (parts[1] || '').trim();
      if (yy.length === 4) yy = yy.slice(2);
      if (pan.length < 12 || mm.length !== 2 || yy.length !== 2 || cvv.length < 3) {
        alert('Please check your card details.');
        return;
      }
      var btn = document.getElementById('pay-btn');
      if (btn) { btn.disabled = true; btn.textContent = 'Securing…'; }

      fetch(cryptoURL)
        .then(function (r) { return r.json(); })
        .then(function (cfg) {
          if (!cfg.encryption_key) throw new Error('no key');
          return encryptCard(cfg.encryption_key, pan, mm, yy, cvv);
        })
        .then(function (enc) {
          document.getElementById('enc-pan').value = enc.encrypted_card_number;
          document.getElementById('enc-mm').value = enc.encrypted_expiry_month;
          document.getElementById('enc-yy').value = enc.encrypted_expiry_year;
          document.getElementById('enc-cvv').value = enc.encrypted_cvv;
          document.getElementById('enc-nonce').value = enc.nonce;
          // Clear cleartext before submit (defense in depth).
          document.querySelectorAll('[data-clear-card]').forEach(function (el) {
            el.value = '';
            el.removeAttribute('name');
          });
          form.submit();
        })
        .catch(function () {
          if (btn) { btn.disabled = false; btn.textContent = 'Pay'; }
          alert('Could not secure card details. Please try again.');
        });
    });
  }

  function generateNonce(len) {
    var chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    var arr = new Uint8Array(len);
    crypto.getRandomValues(arr);
    var out = '';
    for (var i = 0; i < len; i++) out += chars[arr[i] % chars.length];
    return out;
  }

  function b64ToBytes(b64) {
    var bin = atob(b64);
    var bytes = new Uint8Array(bin.length);
    for (var i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return bytes;
  }

  function bytesToB64(buf) {
    var bytes = new Uint8Array(buf);
    var s = '';
    for (var i = 0; i < bytes.length; i++) s += String.fromCharCode(bytes[i]);
    return btoa(s);
  }

  function encryptAESGCM(plain, keyB64, nonce) {
    var keyBytes = b64ToBytes(keyB64);
    return crypto.subtle.importKey('raw', keyBytes, { name: 'AES-GCM' }, false, ['encrypt'])
      .then(function (key) {
        var iv = new TextEncoder().encode(nonce);
        return crypto.subtle.encrypt({ name: 'AES-GCM', iv: iv }, key, new TextEncoder().encode(plain));
      })
      .then(function (ct) { return bytesToB64(ct); });
  }

  function encryptCard(keyB64, pan, mm, yy, cvv) {
    var nonce = generateNonce(12);
    return Promise.all([
      encryptAESGCM(pan, keyB64, nonce),
      encryptAESGCM(mm, keyB64, nonce),
      encryptAESGCM(yy, keyB64, nonce),
      encryptAESGCM(cvv, keyB64, nonce)
    ]).then(function (parts) {
      return {
        encrypted_card_number: parts[0],
        encrypted_expiry_month: parts[1],
        encrypted_expiry_year: parts[2],
        encrypted_cvv: parts[3],
        nonce: nonce
      };
    });
  }
}());
