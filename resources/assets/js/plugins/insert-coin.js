(function($) {
  $.fn.insertCoinBtn = function(options) {
    console.log('this:', this)
    console.log('options:', options)

    this.each(function(this) {
      console.log(this)
    });
  }
})(window.$)
