<template>
    <div>
        <h3>Insert Coin</h3>
        <button @click.prevent="addPayment">Click Me</button>
    </div>
</template>

<script>
define(function () {
    var $flare = window.$flare;

    return {
        props: ['flareView'],
        template: template,
        mounted: function () {
            console.log('params: ', this.$route.params);
        },
        methods: {
            addPayment: function () {
                var path =
                    '<% .Helpers.UrlForRoute "payment:received" "optname" "OPTNAME" "amount" "AMOUNT"  %>';
                path = path.replace('OPTNAME', 'wired-coinslot');
                path = path.replace('AMOUNT', 1);
                $flare.http
                    .post(path)
                    .then(function (data) {
                        console.log(data);
                    })
                    .catch(function (e) {
                        console.error(e);
                    });
            }
        }
    };
});
</script>
