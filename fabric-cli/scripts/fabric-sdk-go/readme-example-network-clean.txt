#
# The process of cleaning up example network artifacts is not yet automated.
#
# Step one: remove example network docker containers.
#
#   Note that the lines below stops and removes ALL conatiners from your
#   system, which may not be what you really want if you are running
#   containers not related to the example network.
#
#       $ docker stop $(docker ps -a -q)
#       $ docker rm $(docker ps -a -q)
#
# Step two: remove example network docker images.
#
#   For repeated testing with bringing the example network up and down,
#   you typically don't remove HLF images as they need time to download
#   and recreate.
#
#   If you are making changes to already deployed and instantiated
#   example_cc.go, you have two options:
#
#   1) deploy the new version of chaincode (specify the next version in
#      command line) and upgrade the chaincode to the new version
#
#   2) delete the chaincode images so you can start deploying/upgrading
#      from v0 again:
#
#       $ docker images | grep examplecc
#       $ docker rmi <examplecc image IDs>

