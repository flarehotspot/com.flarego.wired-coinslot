(function($) {
  $.fn.insertCoinBtn = function(options) {
    console.log('this:', this)
    console.log('options:', options)

    this.each(function(index) {
      $(this).on('click', function(){
        var url = $(this).attr('data-post-url')
        $.ajax({
          method: 'POST',
          url: url,
          success: function(data) {
            console.log('data:', data)
          },
        })
      })
    });
  }
})(window.$)
