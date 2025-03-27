function envVNC(vncID) {
    const RFB = novnc.default
    
    const password = prompt("Enter VNC Password", "");

    // https://github.com/TigerVNC/tigervnc/blob/7938452425de9c73beaa937be886921336aa7795/unix/xserver/hw/vnc/xvnc.c#L751
    const vncRandRWidths = [ 1920, 1920, 1600, 1680, 1400, 1360, 1280, 1280, 1280, 1280, 1024, 800, 640 ];
    const vncRandRHeights = [ 1200, 1080, 1200, 1050, 1050,  768, 1024,  960,  800,  720,  768, 600, 480 ];

    var resolutions = []
    for (ix in vncRandRWidths) {
      resolutions.push(`${vncRandRWidths[ix]}x${vncRandRHeights[ix]}`);
    }

    var geometry = ''
    while (resolutions.indexOf(geometry) == -1) {
      geometry = prompt(`Choose display resolution\n${resolutions.join(', ')}`, "1920x1080");
      // User clicked cancel
      if (geometry == null) {
        return;
      }
    }
    const geometryParts = geometry.split("x", 2);
    const width = geometryParts[0];
    const height = geometryParts[1];

    const vncHost = document.getElementById(vncID)
    const proto = window.location.protocol == "http:" ? "ws" : "wss";
    const host = window.location.host;
    const path = '/api/v1/vnc';
    // const query = `?width=${screen.width}&height=${screen.height}`
    const query = `?width=${width}&height=${height}`
    const wsURL = `${proto}://${host}${path}${query}`;
    const vncArgs = { credentials: { password: password } };

    rfb = new RFB(vncHost, wsURL, vncArgs);

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
