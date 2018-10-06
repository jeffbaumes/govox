from concurrent import futures
import time

import grpc

import govox_pb2
import govox_pb2_grpc

_ONE_DAY_IN_SECONDS = 60 * 60 * 24
pb2 = govox_pb2
pb2_grpc = govox_pb2_grpc
class BuildorbPlugin():
    class blocks():
        AIR = pb2.AIR
        GRASS = pb2.GRASS
        DIRT = pb2.DIRT
        STONE = pb2.STONE
        MOON = pb2.MOON
        ASTEROID = pb2.ASTEROID
        SUN = pb2.SUN
        BLUE_BLOCK = pb2.BLUE_BLOCK
        BLUE_SAND = pb2.BLUE_SAND
        PURPLE_BLOCK = pb2.PURPLE_BLOCK
        PURPLE_SAND = pb2.PURPLE_SAND
        RED_BLOCK = pb2.RED_BLOCK
        RED_SAND = pb2.RED_SAND
        YELLOW_BLOCK = pb2.YELLOW_BLOCK
        YELLOW_SAND = pb2.YELLOW_SAND
        WATER = pb2.WATER

    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:50051')
        self.stub = govox_pb2_grpc.GovoxStub(self.channel)
        self.server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    def addPlanetGen(self, planetgen):
        govox_pb2_grpc.add_GeneratorServicer_to_server(planetgen(), self.server)
    def getPlanets(self):
        return self.stub.GetPlanets(govox_pb2.GetPlanetsRequest())
    def serve(self):
        self.server.add_insecure_port('[::]:50052')
        self.server.start()
        try:
            while True:
                time.sleep(_ONE_DAY_IN_SECONDS)
        except KeyboardInterrupt:
            self.server.stop(0)
