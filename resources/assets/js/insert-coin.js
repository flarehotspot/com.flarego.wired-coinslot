$(document).ready(function () {
  var evt = new EventSource("/sse/events");

  evt.addEventListener("payment:received", function (res) {
    var data = JSON.parse(res.data);
    console.log(data);
  });

  evt.onerror = function (res) {
    console.error(res);
  };
});
