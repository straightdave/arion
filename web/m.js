function toggleInvokePanel(element) {
    console.log("clicked endpoint name bar: " + element.innerText)
    var id = "invokepanel" + element.innerText

    if (typeof(Storage) !== "undefined") {
        var svcName = document.getElementById("svctitle").innerText
        var methodName = element.innerText
        var key = svcName + "#" + methodName
        console.log("loading stored request of " + key)
        var content = localStorage.getItem(key)
        if (content !== "") {
            document.getElementById("invokebody" + methodName).value = content
        }
        else {
            console.log("no stored content for " + key)
        }
    }
    document.getElementById(id).classList.toggle("invisible")
}

function invokeCall(element) {
    var rawStr = "" + element.id
    console.log("clicked invoke button: " + rawStr)
    if (!rawStr.startsWith("invokebutton")) {
        console.log("err: button has no id startsWith 'invoke'")
        return
    }
    var methodName = rawStr.substr(12)
    var reqBody = document.getElementById("invokebody"+methodName).value

    if (typeof(Storage) !== "undefined") {
        var svcName = document.getElementById("svctitle").innerText
        var key = svcName + "#" + methodName
        console.log("Saving req body to " + key)
        localStorage.setItem(key, reqBody)
    }

    callEndpoint(methodName, reqBody)
}

function callEndpoint(name, body) {
    console.log("call endpoint. name="+name+", body="+body)
    var arionPort = document.getElementById("arion-port").value
    var rawAddr = document.getElementById("in-hostaddress").value
    if (rawAddr.trim() === "") {rawAddr = "0.0.0.0:8087"}
    var encodedAddr = encodeURIComponent(rawAddr)
    var url = "http://localhost:"+arionPort+"/call?e="+name+"&h="+rawAddr+"&format=json"
    console.log("calling url: "+url)
    var outputElement = document.getElementById("invokeresult"+name)
    simpleAjaxCall("POST", url, body, function(resp) {
        outputElement.innerHTML = "<pre>"+resp+"</pre>"
    })
}

function textareaKeyup(event, textarea) {
    event.preventDefault()
    if (event.keyCode === 13) {
        // a new line, try to resize (height)
        var lines = textarea.value.split('\n')
        var h = 18*lines.length
        if (h >= 200) {
            textarea.style.height = h + "px"
        }
    }
}

function simpleAjaxCall(method, url, body, cbOK, cbErr) {
    var xhr = new XMLHttpRequest()
    xhr.open(method, url, true)
    xhr.setRequestHeader("Content-type", "application/json")
    xhr.onload = function(e) {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                console.log("Ajax OK")
                if (cbOK) {cbOK(xhr.responseText)}
            } else {
                console.log("Ajax not OK: " + xhr.statusText)
                if (cbErr) {cbErr(xhr.statusText)}
            }
        } else {console.log("XMLHttpRequest not ready")}
    }
    xhr.onerror = function(e) {
        console.log("Ajax error: " + xhr.statusText)
        if (cbErr) {cbErr(xhr.statusText)}
    }
    xhr.send(body)
}
