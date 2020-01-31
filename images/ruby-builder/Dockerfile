# Copyright 2020 Pusher Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#############################################################################
###
### RUBY BUILDER - From quay.io/pusher/builder with Rubyg tooling for
###                  building Ruby based projects
###
#############################################################################

ARG VERSION=latest
FROM quay.io/pusher/builder:${VERSION}

# add env we can debug with the image name:tag
ARG IMAGE_ARG
ENV IMAGE=${IMAGE_ARG}

# Switch to root user to perform installation
USER root

#################
###
### Install Ruby
###
#################

# Taken from docker-library/buildpack-deps:stretch-curl docker image

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    netbase \
    wget \
  && rm -rf /var/lib/apt/lists/*

RUN set -ex; \
  if ! command -v gpg > /dev/null; then \
    apt-get update; \
    apt-get install -y --no-install-recommends \
      gnupg \
      dirmngr \
    ; \
    rm -rf /var/lib/apt/lists/*; \
  fi

# Taken from docker-library/buildpack-deps:stretch-scm docker image

# procps is very common in build systems, and is a reasonably small package
RUN apt-get update && apt-get install -y --no-install-recommends \
    bzr \
    git \
    mercurial \
    openssh-client \
    subversion \
    \
    procps \
  && rm -rf /var/lib/apt/lists/*

# Taken from docker-library/buildpack-deps:stretch docker image

RUN set -ex; \
  apt-get update; \
  apt-get install -y --no-install-recommends \
    autoconf \
    automake \
    bzip2 \
    dpkg-dev \
    file \
    g++ \
    gcc \
    imagemagick \
    libbz2-dev \
    libc6-dev \
    libcurl4-openssl-dev \
    libdb-dev \
    libevent-dev \
    libffi-dev \
    libgdbm-dev \
    libglib2.0-dev \
    libgmp-dev \
    libjpeg-dev \
    libkrb5-dev \
    liblzma-dev \
    libmagickcore-dev \
    libmagickwand-dev \
    libmaxminddb-dev \
    libncurses5-dev \
    libncursesw5-dev \
    libpng-dev \
    libpq-dev \
    libreadline-dev \
    libsqlite3-dev \
    libssl-dev \
    libtool \
    libwebp-dev \
    libxml2-dev \
    libxslt-dev \
    libyaml-dev \
    make \
    patch \
    unzip \
    xz-utils \
    zlib1g-dev \
    \
# https://lists.debian.org/debian-devel-announce/2016/09/msg00000.html
    $( \
# if we use just "apt-cache show" here, it returns zero because "Can't select versions from package 'libmysqlclient-dev' as it is purely virtual", hence the pipe to grep
      if apt-cache show 'default-libmysqlclient-dev' 2>/dev/null | grep -q '^Version:'; then \
        echo 'default-libmysqlclient-dev'; \
      else \
        echo 'libmysqlclient-dev'; \
      fi \
    ) \
  ; \
  rm -rf /var/lib/apt/lists/*

# Taken from docker-library/ruby docker image
# thanks deivid-rodriguez & huacnlee

# skip installing gem documentation
RUN set -eux; \
  mkdir -p /usr/local/etc; \
  { \
    echo 'install: --no-document'; \
    echo 'update: --no-document'; \
  } >> /usr/local/etc/gemrc

ENV RUBY_MAJOR 2.6
ENV RUBY_VERSION 2.6.5
ENV RUBY_DOWNLOAD_SHA256 d5d6da717fd48524596f9b78ac5a2eeb9691753da5c06923a6c31190abe01a62

# some of ruby's build scripts are written in ruby
#   we purge system ruby later to make sure our final image uses what we just built
RUN set -eux; \
  \
  savedAptMark="$(apt-mark showmanual)"; \
  apt-get update; \
  apt-get install -y --no-install-recommends \
    bison \
    dpkg-dev \
    libgdbm-dev \
    ruby \
  ; \
  rm -rf /var/lib/apt/lists/*; \
  \
  wget -O ruby.tar.xz "https://cache.ruby-lang.org/pub/ruby/${RUBY_MAJOR%-rc}/ruby-$RUBY_VERSION.tar.xz"; \
  echo "$RUBY_DOWNLOAD_SHA256 *ruby.tar.xz" | sha256sum --check --strict; \
  \
  mkdir -p /usr/src/ruby; \
  tar -xJf ruby.tar.xz -C /usr/src/ruby --strip-components=1; \
  rm ruby.tar.xz; \
  \
  cd /usr/src/ruby; \
  \
# hack in "ENABLE_PATH_CHECK" disabling to suppress:
#   warning: Insecure world writable dir
  { \
    echo '#define ENABLE_PATH_CHECK 0'; \
    echo; \
    cat file.c; \
  } > file.c.new; \
  mv file.c.new file.c; \
  \
  autoconf; \
  gnuArch="$(dpkg-architecture --query DEB_BUILD_GNU_TYPE)"; \
  ./configure \
    --build="$gnuArch" \
    --disable-install-doc \
    --enable-shared \
  ; \
  make -j "$(nproc)"; \
  make install; \
  \
  apt-mark auto '.*' > /dev/null; \
  apt-mark manual $savedAptMark > /dev/null; \
  find /usr/local -type f -executable -not \( -name '*tkinter*' \) -exec ldd '{}' ';' \
    | awk '/=>/ { print $(NF-1) }' \
    | sort -u \
    | xargs -r dpkg-query --search \
    | cut -d: -f1 \
    | sort -u \
    | xargs -r apt-mark manual \
  ; \
  apt-get purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false; \
  \
  cd /; \
  rm -r /usr/src/ruby; \
# verify we have no "ruby" packages installed
  ! dpkg -l | grep -i ruby; \
  [ "$(command -v ruby)" = '/usr/local/bin/ruby' ]; \
# rough smoke test
  ruby --version; \
  gem --version; \
  bundle --version

# don't create ".bundle" in all our apps
ENV GEM_HOME /usr/local/bundle
ENV BUNDLE_SILENCE_ROOT_WARNING=1 \
  BUNDLE_APP_CONFIG="$GEM_HOME"
ENV PATH $GEM_HOME/bin:$PATH
# adjust permissions of a few directories for running "gem install" as an arbitrary user
RUN mkdir -p "$GEM_HOME" && chmod 777 "$GEM_HOME"

# Switch back to prow user for publish
USER prow
