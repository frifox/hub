let socket = new Socket("ws://" + location.host + "/ws");
let hub = new Hub()

function Socket(url) {
    let WS = init()
    function reconnect (){
        WS = init()
    }
    function send(emit) {
        WS.send(JSON.stringify(emit))
    }

    function init() {
        console.log("Socket Connecting");

        const ws = new WebSocket(url)
        ws.onopen = () => {
            console.log("Socket Connected");
        };
        ws.onclose = event => {
            console.log("Socket Closed: ", event);
            hub.clear()
            setTimeout(function() {
                reconnect()
            }, 5000)
        };
        ws.onerror = error => {
            console.log("Socket Error: ", error);
            socket.close();
        };

        ws.onmessage = message => {
            const data = JSON.parse(message.data)
            console.log("[FromWS]", data, data.Data)
            hub.handleEmit(data)
        };

        return ws
    }

    return {
        send,
    }
}

function Hub() {
    let projectors = {} // []Projector
    let decoders = {};  // []Decoder
    let encoders = {};  // []string

    function add(DeviceType, DeviceName) {
        switch(DeviceType) {
            case "projector":
                projectors[DeviceName] = new Projector(DeviceName)
                document.getElementById("projectors").appendChild(projectors[DeviceName].element)
                return true
            case "decoder":
                decoders[DeviceName] = new Decoder(DeviceName)
                document.getElementById("decoders").appendChild(decoders[DeviceName].element)
                return true
            case "encoder":
                encoders[DeviceName] = DeviceName
                for (const decoder in decoders) {
                    decoders[decoder].add(new Encoder(DeviceName))
                }
                return true
            default:
                console.log("hub.add() unhandled DeviceType:", DeviceType, DeviceName)
                return false
        }
    }

    function get(DeviceType, DeviceName) {
        switch (DeviceType) {
            case "projector":
                return projectors[DeviceName]
            case "decoder":
                return decoders[DeviceName]
            default:
                return null
        }
    }

    function clear() {
        for (const projector in projectors) {
            projectors[projector].element.remove()
        }
        for (const decoder in decoders) {
            decoders[decoder].element.remove()
        }
    }

    function handleEmit(emit) {
        switch (emit.Command) {
            case "Init":
                let ok = add(emit.DeviceType, emit.DeviceName)
                if (!ok) {
                    return
                }

                // request status refresh
                socket.send({
                    "DeviceType": emit.DeviceType,
                    "DeviceName": emit.DeviceName,
                    "Command": "Refresh"
                })
                break
            default:
                let device = get(emit.DeviceType, emit.DeviceName)

                if (device == null) {
                    console.log("Hub 404", emit)
                    return
                }
                device.handleEmit(emit)
                break
        }
    }

    return {
        handleEmit,
        clear
    }
}

function Projector(name) {
    const button = document.createElement("button")
    button.setAttribute("class", "btn btn-dark")
    button.textContent = name
    button.onclick = function() {
        socket.send({
            "DeviceType": "projector",
            "DeviceName": name,
            "Command": "Refresh",
        })
    }

    const Freeze = new ButtonStatus("Freeze")
    Freeze.element.onclick = function() {
        Freeze.emitToggle("projector", name)
    }

    const Blank = new ButtonStatus("Blank")
    Blank.element.onclick = function() {
        Blank.emitToggle("projector", name)
    }

    const Power = new ButtonStatus("Power")
    Power.element.onclick = function() {
        Power.emitToggle("projector", name)
    }

    const div = document.createElement("div")
    div.appendChild(button)
    div.appendChild(Freeze.element)
    div.appendChild(Blank.element)
    div.appendChild(Power.element)

    function handleEmit(emit) {
        const command = emit.Command
        const status = emit.Data.Status

        switch (command) {
            case "Freeze":
                Freeze.handleStatus(status)
                break
            case "Blank":
                Blank.handleStatus(status)
                break
            case "Power":
                Power.handleStatus(status)
                break
            default:
                console.log("[Projector] unhandled Command", Power)
                break
        }
    }

    return {
        element: div,
        handleEmit,
    }
}

function ButtonStatus(name) {
    const button = document.createElement("button")
    button.setAttribute("class", "btn btn-light")
    button.setAttribute("disabled", "disabled")
    button.textContent = name

    let status = ""
    function handleStatus(value) {
        button.removeAttribute("disabled")
        switch (value) {
            case "on":
                status = value
                button.setAttribute("class", "btn btn-success")
                break
            case "off":
                status = value
                button.setAttribute("class", "btn btn-secondary")
                break
            default:
                button.setAttribute("class", "btn btn-warning")
                console.log("[Button:" + name + "] unhandled status: " + value)
                break
        }
    }
    function emitToggle(DeviceType, DeviceName) {
        button.setAttribute("disabled", "disabled")
        let emit = {
            "DeviceType": DeviceType,
            "DeviceName": DeviceName,
            "Command": name,
            "Status": "",
        }
        switch (status) {
            case "on":
                emit.Status = "off"
                break
            case "off":
                emit.Status = "on"
                break
            default:
                emit.Status = "" // will trigger related Inquiry command
                break
        }
        socket.send(emit)
        console.log("[ToWS]", emit)
    }
    return {
        element: button,
        handleStatus,
        emitToggle,
    }
}

function Decoder(name) {
    const div = document.createElement("div")
    div.setAttribute("DeviceName", name)

    const button = document.createElement("button")
    button.setAttribute("class", "btn btn-dark")
    button.textContent = name
    button.onclick = function() {
        console.log("Decoder TODO:", name)
    }
    div.appendChild(button)

    function add(encoder) {
        encoder.element.onclick = function() {
            socket.send({
                "DeviceType": "decoder",
                "DeviceName": name,
                "Command": "ConnectTo",
                "Encoder": encoder.name
            })
            console.log("Encoder:", encoder.name, "Decoder:", name)
        }
        div.appendChild(encoder.element)
    }

    function handleEmit(emit) {
        console.log("[Decoder]", emit)
    }

    return {
        element: div,
        handleEmit,
        add
    }
}
function Encoder(name) {
    const button = document.createElement("button")
    button.textContent = name
    button.setAttribute("class", "btn btn-light")

    return {
        name,
        element: button,
    }
}