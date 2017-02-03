#!/usr/bin/env python
#
# import socket
#
# TCP_IP = '127.0.0.1'
# TCP_PORT = 1111
# BUFFER_SIZE = 20  # Normally 1024, but we want fast response
#
# s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
# s.bind((TCP_IP, TCP_PORT))
# s.listen(1)
#
# conn, addr = s.accept()
# print 'Connection address:', addr
# while 1:
#     data = conn.recv(BUFFER_SIZE)
#     if not data: break
#     print "received data:", data
#     conn.send(data)  # echo
# conn.close()

import socket
import sys

# Create a TCP/IP socket
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

# Bind the socket to the address given on the command line
server_name = socket.gethostname()
server_address = (server_name, 1111)
print >>sys.stderr, 'starting up on %s port %s' % server_address
sock.bind(server_address)
sock.listen(1)

while True:
    print >>sys.stderr, 'waiting for a connection'
    connection, client_address = sock.accept()
    try:
        print >>sys.stderr, 'client connected:', client_address
        while True:
            data = connection.recv(16)
            print >>sys.stderr, 'received "%s"' % data
            if data:
                connection.sendall(data)
            else:
                break
    finally:
        connection.close()