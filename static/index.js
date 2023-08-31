const ws = new WebSocket("ws://localhost:3000/ws")

const appendMessage = (from, message) => {
    const html = `<div class="alert alert-primary mb-2">
            <div class="d-flex align-items-center justify-content-between">
                <h6 class="fw-bold">@${from}</h6>
                <span class="text-muted"></span>
            </div>
            <span>${message}</span>
        </div>`

    document.getElementById("messages-container").innerHTML += html
}

const appendNotification = (message) => {
    const html = `<div class="alert alert-secondary mb-2">
        <div class="card-body">
            <div class="d-flex align-items-center justify-content-between">
                <span>${message}</span>
                <span class="text-muted"></span>
            </div>
        </div>
    </div>`

    document.getElementById("messages-container").innerHTML += html
}

const appendError = (message) => {
    const html = `<div class="alert alert-danger mb-2">
        <div class="card-body">
            <div class="d-flex align-items-center justify-content-between">
                <span>${message}</span>
                <span class="text-muted"></span>
            </div>
        </div>
    </div>`

    document.getElementById("messages-container").innerHTML += html
}

ws.onopen = () => {
    console.log("connection opened")
}

ws.onerror = (e) => {
    appendError("Something went wrong")
}

ws.onclose = () => {
    appendError("Connection closed")
}

ws.onmessage = (e) => {
    const msg = JSON.parse(e.data)

    if (msg.type === "message") {
        appendMessage(msg.from, msg.content)
    } else if (msg.type === "notification") {
        appendNotification(msg.content)
    }

    document.getElementById("messages-container").scrollTo(0, document.body.scrollHeight);
}
