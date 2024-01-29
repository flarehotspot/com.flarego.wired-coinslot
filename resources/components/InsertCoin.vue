<template>
  <h3>Insert Coin</h3>
</template>

<script>
define(function () {
  return {
    props: ['flareView'],
    template: template,
    mounted: function () {
      console.log('params: ', this.$route.params);
    }
  };
});
</script>
