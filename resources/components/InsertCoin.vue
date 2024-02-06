<template>
    <div>
        <h3>Insert Coin</h3>
        <h4>Payment For: {{ data.purchase_name }}</h4>
        <p>Total Payment: {{ data.purchase_state.total_payment }}</p>
        <button @click.prevent="addPayment">Click Me</button>
        <button @click.prevent="donePayment">Done Paying</button>
    </div>
</template>

<script>
define(function () {
    var $flare = window.$flare;

    return {
        props: ['flareView'],
        template: template,
        computed: {
            data: {
                get: function () {
                    var data = this.flareView.data;
                    if (!data.purchase_state) data.purchase_state = {};
                    return data;
                },
                deep: true
            }
        },
        methods: {
            addPayment: function () {
                var self = this;
                var path =
                    '<% .Helpers.UrlForRoute "payment:received" "optname" "OPTNAME" "amount" "AMOUNT"  %>';
                path = path.replace('OPTNAME', self.flareView.data.optname);
                path = path.replace('AMOUNT', 1);
                $flare.http
                    .post(path)
                    .then(function (data) {
                        console.log(data);
                        self.flareView.data = data;
                    })
                    .catch(function (e) {
                        console.error(e);
                    });
            },
            donePayment: function () {
                var path = '<% .Helpers.UrlForRoute "payment:done" %>';
                $flare.http.post(path);
            }
        }
    };
});
</script>
