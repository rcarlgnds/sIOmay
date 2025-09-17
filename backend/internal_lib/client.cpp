#include "internal_lib.hpp"
void startClient(std::string listenAddr, int port){
    asio::io_context io_context;
    
    try{
        UdpMultiCastClient client(io_context, listenAddr, port);
        client.listen_loop();

    }catch (std::exception& e) {
        std::cerr << "Exception: " << e.what() << std::endl;
    }
}
extern "C" void startClientC(const char* listenAddr, int port) {
    std::string addr(listenAddr);
    startClient(addr, port);
}