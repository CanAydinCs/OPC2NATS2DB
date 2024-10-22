from opcua import Server
from random import randint
import datetime
import time

# start on local
IP_ADRESS = "0.0.0.0"

server = Server()
url = "opc.tcp://" + str(IP_ADRESS) + ":4840"
server.set_endpoint(url)

name = "OPCUA_SIMULATION_SERVER"
addspace = server.register_namespace(name)

node = server.get_objects_node()

Param = node.add_object(addspace, "Parameters")

Temp = Param.add_variable(addspace, "Tempature", 0)
Press = Param.add_variable(addspace, "Pressure", 0)
Time = Param.add_variable(addspace, "Time", 0)

Temp.set_writable()
Press.set_writable()
Time.set_writable()

server.start()
print("Server started at {}".format(url))

while True:
	Temperature = randint(10,50)
	Pressure = randint(200,900)
	Tim = datetime.datetime.now()

	print(Temperature, Pressure, Tim)

	Temp.set_value(Temperature)
	Press.set_value(Pressure)
	Time.set_value(Tim)
	
	time.sleep(2)