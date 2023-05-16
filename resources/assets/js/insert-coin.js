$(document).ready(function () {
  var evt = new EventSource("/sse/events");
  var checkbox = $("#use-wallet-bal-checkbox");
  var walletBal = $("#wallet-bal");

  function init() {
    $.ajax();
  }

  function procData(data) {
    if (data.wallet_debit > 0) {
      checkbox.attr("checked", "checked");
    } else {
      checkbox.removeAttr("checked");
    }

    walletBal.text(data.wallet_avail_bal + "/" + data.wallet_bal);
  }

  evt.addEventListener("payment:received", function (res) {
    var data = JSON.parse(res.data);
    alert("Payment recieved: " + data.amount);
  });

  evt.onerror = function (res) {
    console.error(res);
  };

  checkbox.change(function () {
    var checked = this.checked;
    var url = window.USE_WALLET_URL;
    var amount = checked ? window.WALLET_BAL;

    $.ajax({
      method: "GET",
      url: url,
      success: function (data) {
        console.log(data);
        walletBal.text(data.wallet_bal);

        if (checked) {
          $(this).attr("checked", "checked");
        } else {
          $(this).removeAttr("checked");
        }
      },
      fail: function (err) {
        console.error(err);
        // walletBal.text("0");
      },
    });
  });
});
