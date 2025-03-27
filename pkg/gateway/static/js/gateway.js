function envVNC(vncID) {
    const RFB = novnc.default
    
    const password = prompt("Enter VNC Password", "");

    const vncHost = document.getElementById(vncID)
    rfb = new RFB(vncHost, `${window.location.protocol == "http:" ? "ws" : "wss"}://${window.location.host}/api/v1/vnc`, { 
      credentials: { password: password },
    });

    rfb.addEventListener("connect",  function() {
      alert('Connected');
    });
    rfb.addEventListener("disconnect", function() {
      alert('Disconnected from VNC, refresh to reconnect');
    });
    rfb.addEventListener("credentialsrequired", function() {
      alert('Incorrect Credentials');
    });
    rfb.addEventListener("desktopname", function(e) {
      alert(`Desktop name changed to, ${e.detail.name}`);
    });
    rfb.scaleViewport = true;
}
