// Insert-coin page: subscribes to the server's coin SSE stream and live-updates
// the running balance. Shows a "counting payment" cue on the first pulse of a
// coin (before its amount resolves), and runs an idle countdown that the server
// resets on every real pulse/coin; when it elapses the page auto-finalizes —
// submitting the accumulated payment if any was inserted, otherwise cancelling.
// ES5 only; jQuery is provided globally as window.$.
(function ($) {
  function init() {
    var $page = $('#insert-coin-page');
    if (!$page.length) {
      return;
    }

    var eventsUrl = $page.attr('data-events-url');
    var mockUrl = $page.attr('data-mock-url');
    var doneUrl = $page.attr('data-done-url');
    var cancelUrl = $page.attr('data-cancel-url');
    var countingText = $page.attr('data-counting-text') || 'Counting payment...';
    var receivedText = $page.attr('data-received-text') || 'Received {amount}';

    var $total = $('#coin-total');
    var $status = $('#coin-status');
    var $countdown = $('#coin-countdown');
    var $done = $('#done-btn');
    var $mock = $('#mock-coin-btn');

    // Running state. `total` mirrors the server's authoritative balance and
    // decides which way the timeout finalizes. `finalizing` guards the one-shot
    // navigation so the server timeout event and the local countdown can't both
    // fire it.
    var total = 0;
    var timeoutSecs = 0;
    var remaining = 0;
    var countdownTimer = null;
    var finalizing = false;
    var es = null;

    function setSufficient(ok) {
      if (ok) {
        $done.removeClass('disabled');
      } else {
        $done.addClass('disabled');
      }
    }
    setSufficient(false);

    function renderCountdown(secs) {
      $countdown.text(secs < 0 ? 0 : secs);
    }

    // Restart the idle countdown from the latest server-provided timeout. Called
    // on every activity event (first pulse, resolved coin) so any coin/pulse
    // gives the user a fresh full window.
    function resetCountdown() {
      if (timeoutSecs <= 0) {
        return;
      }
      remaining = timeoutSecs;
      renderCountdown(remaining);
      if (!countdownTimer) {
        countdownTimer = window.setInterval(tick, 1000);
      }
    }

    function tick() {
      if (finalizing) {
        return;
      }
      remaining -= 1;
      if (remaining <= 0) {
        renderCountdown(0);
        finalize();
        return;
      }
      renderCountdown(remaining);
    }

    // Auto-finalize when the countdown elapses: submit the accumulated payment if
    // any coins were inserted, otherwise cancel. Both targets are plain GET
    // routes that tear down the session and redirect (done -> purchase callback,
    // cancel -> portal).
    function finalize() {
      if (finalizing) {
        return;
      }
      finalizing = true;
      if (countdownTimer) {
        window.clearInterval(countdownTimer);
        countdownTimer = null;
      }
      if (es) {
        es.close();
      }
      window.location.href = total > 0 ? doneUrl : cancelUrl;
    }

    // Block the Done link until enough has been paid.
    $done.on('click', function (e) {
      if ($done.hasClass('disabled')) {
        e.preventDefault();
      }
    });

    // Dev helper: ask the server to inject one synthetic coin pulse.
    $mock.on('click', function () {
      $.ajax({ method: 'POST', url: mockUrl });
    });

    if (typeof window.EventSource === 'undefined') {
      $status.text('Live updates are not supported by this browser.');
      return;
    }

    es = new EventSource(eventsUrl);
    es.onmessage = function (msg) {
      if (finalizing) {
        return;
      }
      var data;
      try {
        data = JSON.parse(msg.data);
      } catch (err) {
        return;
      }

      // Adopt the latest timeout and (re)start the countdown. The initial
      // snapshot carries it too, so the countdown begins as soon as the page is
      // live — even before the first coin.
      if (data.timeout_secs > 0) {
        timeoutSecs = data.timeout_secs;
        if (!countdownTimer) {
          resetCountdown();
        }
      }

      // Leading-edge cue: pulses are arriving but the amount isn't resolved yet.
      // Don't touch the balance; just show the cue and refresh the countdown.
      if (data.counting) {
        $status.text(countingText);
        resetCountdown();
        return;
      }

      total = Number(data.total);
      $total.text(total.toFixed(2));
      setSufficient(!!data.sufficient);

      // A resolved coin: show its value and give the user a fresh window.
      if (data.last_coin > 0) {
        $status.text(receivedText.replace('{amount}', Number(data.last_coin).toFixed(2)));
        resetCountdown();
      }
    };
    es.onerror = function () {
      if (!finalizing) {
        $status.text('Reconnecting...');
      }
    };
  }

  $(document).ready(init);
})(window.$);
