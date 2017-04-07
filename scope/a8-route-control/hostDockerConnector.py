# Socket server that exists on the host machine to get information from the Docker API for the 
# app in the container

import socket,os
import subprocess
import json
from socket import error as SocketError
import errno
script_dir = os.path.dirname(__file__) 
abs_file_path = os.path.join(script_dir, "dockerConnection/hostconnection.sock")
abs_file_path = "/tmp/dockerConnection/hostconnection.sock"
s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

def getDockerIDsWithIPs():
    cmd = ["docker inspect -f '{{.Id}}-{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -q)"]
    p = subprocess.Popen(cmd, stdout=subprocess.PIPE, shell=True)
    (output, err) = p.communicate()
    #print output
    # create a json array of the containerID, IP pairs
    arr = []
    lines = output.splitlines()
    for i,line in enumerate(lines):
        #filter out the container IDs that don't have an IP address
        pair = line.split('-')
        if len(pair) == 2:
            if pair[1] not in [""]:
                #print pair
                # add pair to the array
                obj = {"id":pair[0], "ip":pair[1]}
                arr.append(obj)
    #print json.dumps(arr)
    return json.dumps(arr)

try:
    os.remove(abs_file_path)
except OSError:
    pass
s.bind(abs_file_path)
s.listen(1)

while 1:
    try:
        conn, addr = s.accept()
        while 1:
            data = conn.recv(1024)
            data = getDockerIDsWithIPs()
            data += "\n"
            
            if not data: 
                break
            print data
            conn.send(data)
        conn.close()
    except SocketError as e:
        pass # Handle error here.

