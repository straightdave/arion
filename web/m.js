(function(){
    // test data
    var _data = {
        num: 2,
        svc: [
            {
                name: "svc1",
                endpoints: []
            }
        ]
    }

    // to gen innerHTML
    var gen = function(data) {
        var str = ''





        return "hello world"
    }

    window.onload=function() {
        console.log("page loaded")
        var mainDiv = document.getElementById("main")
        if (mainDiv === null) {
            console.err("get no element of id=main")
            return
        }
        mainDiv.innerHTML = gen(_data)
    }
})()
