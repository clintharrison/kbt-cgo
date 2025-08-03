# The particular version/distro here is not hugely important, but we do need
# a base image with 7zip (not p7zip) at least version 23.00
FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

# Install the general build dependencies as well as the dependencies of the
# koxtoolchain and kindle-sdk repos.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install \
    git \
    build-essential \
    autoconf \
    automake \
    bison \
    flex \
    gawk \
    libtool \
    libtool-bin \
    libncurses-dev \
    curl \
    file \
    git \
    gperf \
    help2man \
    texinfo \
    unzip \
    wget \
    7zip \
    # For the sdk
    curl \
    wget \
    sed \
    libarchive-dev \
    nettle-dev \
    # direnv is required by the VS Code Dev Container extension
    direnv \
    # GTK2.0/Meson tutorial dependencies
    ninja-build \
    libgtk2.0-dev \
    # to embed the main.ui file in the binary
    xxd \
    # for newer Meson than Ubuntu Noble has
    python3 \
    python3-pip \
    python3-setuptools \
    python3-wheel \
    pipx

# Install Go SDK
RUN wget -O /tmp/go.tar.gz https://go.dev/dl/go1.24.4.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm -f /tmp/go.tar.gz && \
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/bash.bashrc

RUN pipx install meson \
    && pipx ensurepath

# crosstool-ng requires heavy opt-in to run the build as root, but it's fine
# here because we're building in a container.
# We also may not need to _build_ anything, if we can fetch a pre-built toolchain.
ENV CT_EXPERIMENTAL=y CT_ALLOW_BUILD_AS_ROOT=y CT_ALLOW_BUILD_AS_ROOT_SURE=y

ARG platform=kindlehf

# Set up the kindle toolchain
COPY setup-toolchain.sh /tmp/setup-toolchain.sh
RUN /tmp/setup-toolchain.sh $platform && \
    rm -f /tmp/setup-toolchain.sh

# This speeds up the firmware download
RUN apt-get -y install aria2

# Set up the kindle-sdk
RUN git clone --recursive --depth=1 \
    https://github.com/clintharrison/kindle-sdk \
    --branch clint/need-5.17plus-bluetooth /opt/kindle-sdk \
    && \
    /opt/kindle-sdk/gen-sdk.sh \
        kindlehf \
        /root/x-tools/arm-kindlehf-linux-gnueabihf

COPY aliases.bashrc /tmp/aliases.bashrc
RUN cat /tmp/aliases.bashrc >> /etc/bash.bashrc && \
    rm -f /tmp/aliases.bashrc

COPY fix-up-sdk.sh /tmp/fix-up-sdk.sh
RUN /tmp/fix-up-sdk.sh && \
    rm -f /tmp/fix-up-sdk.sh