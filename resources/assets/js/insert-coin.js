$(document).ready(function () {
  var evt = new EventSource("/sse/events");
  var checkbox = $("#use-wallet-bal-checkbox");
  var walletBal = $("#wallet-bal");
  var totalAmount = $("#total-amount");
  var url = window.USE_WALLET_URL;

  function procData(data) {
    if (data.WalletDebit > 0) {
      checkbox.attr("checked", "checked");
    } else {
      checkbox.removeAttr("checked");
    }

    walletBal.text(parseFloat(data.WalletAvailBal).toFixed(2) + "/" + parseFloat(data.WalletBal).toFixed(2));
    totalAmount.text(parseFloat(data.PaymentTotal).toFixed(2));
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
    var amount = checked ? window.WalletBal : 0;

    $.ajax({
      method: "GET",
      url: url + "?amount=" + amount,
      success: function (data) {
        console.log(data);
        procData(data);
      },
      fail: function (err) {
        console.error(err);
        // walletBal.text("0");
      },
    });
  });
});
