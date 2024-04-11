#!/bin/sh
if [ ! -z "$SSH_KNOWN_HOSTS_CONTENT" ]; then
  echo "$SSH_KNOWN_HOSTS_CONTENT" > $SSH_KNOWN_HOSTS
fi

exec /root/main
