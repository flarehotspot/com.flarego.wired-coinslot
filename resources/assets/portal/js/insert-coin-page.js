// Insert-coin page: subscribes to the server's coin SSE stream and live-updates
// the running balance. ES5 only; jQuery is provided globally as window.$.
(function ($) {
  function init() {
    var $page = $('#insert-coin-page');
    if (!$page.length) {
      return;
    }

    var eventsUrl = $page.attr('data-events-url');
    var mockUrl = $page.attr('data-mock-url');

    var $total = $('#coin-total');
    var $status = $('#coin-status');
    var $done = $('#done-btn');
    var $mock = $('#mock-coin-btn');

    function setSufficient(ok) {
      if (ok) {
        $done.removeClass('disabled');
      } else {
        $done.addClass('disabled');
      }
    }
    setSufficient(false);

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

    var es = new EventSource(eventsUrl);
    es.onmessage = function (msg) {
      var data;
      try {
        data = JSON.parse(msg.data);
      } catch (err) {
        return;
      }
      $total.text(Number(data.total).toFixed(2));
      setSufficient(!!data.sufficient);
      if (data.last_coin > 0) {
        $status.text('Received ' + Number(data.last_coin).toFixed(2));
      }
    };
    es.onerror = function () {
      $status.text('Reconnecting...');
    };
  }

  $(document).ready(init);
})(window.$);
