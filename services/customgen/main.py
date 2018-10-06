import lib
p = lib.BuildorbPlugin()
class SolidPlanetGenerator(lib.pb2_grpc.GeneratorServicer):
    def CellMaterial(self, request, context):
        return lib.pb2.CellMaterialResponse(cell=lib.pb2.Cell(material=p.blocks.STONE))

if __name__ == '__main__':
    p.addPlanetGen(SolidPlanetGenerator)
    print(p.getPlanets())
    p.serve()
