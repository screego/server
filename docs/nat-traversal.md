# NAT Traversal

In most cases peers cannot directly communicate with each other because of firewalls or other restrictions like NAT.
To work around this issue, WebRTC uses 
[Interactive Connectivity Establishment (ICE)](http://en.wikipedia.org/wiki/Interactive_Connectivity_Establishment).
This is a framework for helping to connect peers.

ICE uses STUN and/or TURN servers to accomplish this.

?> Screego exposes a STUN and TURN server. You don't have to configure this separately.
   By default, user authentication is required for using TURN.

## STUN

[Session Traversal Utilities for NAT (STUN)](http://en.wikipedia.org/wiki/STUN) is used to find
the public / external ip of a peer. This IP is later sent to others to create a direct connection.

When STUN is used, only the connection enstablishment will be done through Screego. The actual video stream will be
directly sent to the other peer and doesn't go through Screego.

While STUN should work for most cases, there are stricter NATs f.ex. 
[Symmetric NATs](https://en.wikipedia.org/wiki/Network_address_translation) 
where it doesn't, then, TURN will be used.

## TURN

[Traversal Using Relays around NAT (TURN)](http://en.wikipedia.org/wiki/TURN) is used to work around Symmetric NATs.
It does it by relaying all data through a TURN server. As relaying will create traffic on the server,
Screego will require user authentication to use the TURN server. This can be configured see [Configuration](config.md).

