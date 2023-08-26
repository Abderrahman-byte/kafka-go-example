const ws = new WebSocket("ws://localhost:3000/ws")

const appendMessage = (message, tag = "info") => {
    const elt = document.createElement("div")
    const p = document.createElement("p")

    p.textContent = message
    elt.className = "alert alert-" + tag +" mb-2"
    elt.appendChild(p)

    document.getElementById("messages").appendChild(elt)
}

ws.onmessage = (e) => {
    appendMessage(e.data, "info")
}

ws.onopen = e => {
    appendMessage("Connection opened", "success")
}

ws.onclose = e => {
    appendMessage("Connection closed", "danger")
}
