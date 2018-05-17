
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
    console.log("calling method: " + methodName)

    // get content to send
    var reqBody = document.getElementById("invokebody"+methodName).value
    console.log("body to send: " + reqBody)

    // call local api to send data
    var response = callEndpoint(methodName, reqBody)
    console.log("response: " + response)

    // render response
    document.getElementById("invokeresult"+methodName).innerText = response

    console.log("done")
}

function callEndpoint(name, body) {
    var xhttp = new XMLHttpRequest();
    // TODO: port
    xhttp.open("POST", "http://localhost:9999/call?e="+name, false);
    xhttp.setRequestHeader("Content-type", "application/json");
    xhttp.send(body);
    return xhttp.responseText;
}
