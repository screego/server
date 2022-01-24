<p align="center">
    <a href="https://screego.net">
        <img src="docs/logo.png" />
    </a>
</p>


<h1 align="center">screego/server</h1>
<p align="center"><i>screen sharing for developers</i></p>

<p align="center">
    <a href="https://github.com/screego/server/actions?query=workflow%3Abuild">
        <img alt="Build Status" src="https://github.com/screego/server/workflows/build/badge.svg">
    </a>
    <a href="https://goreportcard.com/report/github.com/screego/server">
        <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/screego/server">
    </a>
    <a href="https://hub.docker.com/r/screego/server">
        <img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/screego/server.svg">
    </a>
    <a href="https://github.com/screego/server/releases/latest">
        <img alt="latest release" src="https://img.shields.io/github/release/screego/server.svg">
    </a>
</p>

## Intro

In the past I've had some problems sharing my screen with coworkers using
corporate chatting solutions like Microsoft Teams. I wanted to show them some
of my code, but either the stream lagged several seconds behind or the quality
was so poor that my colleagues couldn't read the code. Or both.

That's why I created screego. It allows you to share your screen with good
quality and low latency. Screego is an addition to existing software and 
only helps to share your screen. Nothing else (:.

## Features

* Multi User Screenshare
* Secure transfer via WebRTC
* Low latency / High resolution
* Simple Install via Docker / single binary
* Integrated TURN Server see [NAT Traversal](https://screego.net/#/nat-traversal)

[Demo / Public Instance](https://app.screego.net/) ᛫ [Installation](https://screego.net/#/install) ᛫ [Configuration](https://screego.net/#/config) 

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the
[tags on this repository](https://github.com/screego/server/tags).
