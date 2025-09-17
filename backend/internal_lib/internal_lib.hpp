#pragma once
#include "./network/Multicast.hpp"
#include "./mouse/Mouse.hpp"

struct TrackServer{
    MouseCapture *capture;
    KeyboardCapture *kCapture;
    asio::io_context io_context;
    UdpMulticastServer * server;
    std::string multicast_address;
    int multicast_port;
    

    int startTrackServer();
    TrackServer(std::string multicast_address, int multicast_port);
    
    void sendStopSignal(int ip[4]);
};

void startClient(std::string listenAddr, int port);


#ifdef __cplusplus
extern "C" {
#endif

void startClientC(const char* listenAddr, int port);
void startSiomayServerC();
void sendStopCommandC(int ip[4]);

#ifdef __cplusplus
}
#endif