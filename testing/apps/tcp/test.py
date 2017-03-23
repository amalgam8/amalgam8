#!/usr/bin/env python

import socket
import sys

# Create a TCP/IP socket
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

# Connect the socket to the port on the server given by the caller
server_address = (sys.argv[1], int(sys.argv[2]))
sock.connect(server_address)

try:

    message = 'This is the message.  It will be repeated.'
    sock.sendall(message)

    amount_received = 0
    amount_expected = len(message)
    while amount_received < amount_expected:
        data = sock.recv(16)
        amount_received += len(data)
        print >>sys.stdout, '%s' % data

finally:
    sock.close()
