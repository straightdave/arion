function toggleInvokePanel(element) {
    console.log("clicked endpoint name bar: " + element.innerText)
    var id = "invokepanel" + element.innerText
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
    callEndpoint(methodName, reqBody)
}

function callEndpoint(name, body) {
    console.log("call endpoint. name="+name+", body="+body)
    var arionPort = document.getElementById("arion-port").value
    var rawAddr = document.getElementById("in-hostaddress").value
    if (rawAddr.trim() === "") {rawAddr = "0.0.0.0:8087"}
    var encodedAddr = encodeURIComponent(rawAddr)
    var url = "http://localhost:"+arionPort+"/call?e="+name+"&h="+rawAddr
    console.log("calling url: "+url)
    var outputElement = document.getElementById("invokeresult"+name)
    simpleAjaxCall("POST", url, body, function(resp) {
        outputElement.innerText = resp
    })
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
