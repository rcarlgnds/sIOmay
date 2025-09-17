#include "internal_lib.hpp"
#include <thread>

TrackServer * instance = nullptr;
TrackServer::TrackServer(std::string multicast_address, int multicast_port){
    this->kCapture = KeyboardCapture::GetInstance();
    this->capture = MouseCapture::GetInstance();
    this->server = new UdpMulticastServer(this->io_context, this->multicast_address, this->multicast_port);
 }

int TrackServer::startTrackServer(){

    // start mouse capture
    std::thread poller(PollMouseWindows, std::ref(*capture));
    poller.detach(); // run polling in background

    startHook();
    std::thread t(MessagePump);

    std::thread trackKey(startKeyboardTrack);
    trackKey.detach();

     try {
        // Run the sending loop
        std::thread mouseSend([&]() {
            this->server->send_loop(10, capture);
        });
        mouseSend.detach();
        this->server->send_loop(10, kCapture);

    } catch (std::exception& e) {
        std::cerr << "Exception: " << e.what() << std::endl;
    }

    return 0;
}

void TrackServer::sendStopSignal(int ip[4]){
    uint8_t buf[16] = {};
    formatStopCommandData(buf, 16, ip);
    this->server->send_command(buf, 16);
}


void startSiomayServer(){
    if (instance) {

        return;
    };
    instance = new TrackServer("239.255.0.1", 8080);
    instance->startTrackServer();
}

void sendStopCommand(int ip[4]){
    if (instance){
        instance->sendStopSignal(ip);
    }else{
        throw std::runtime_error("Server not started");
    }

}
extern "C" void startSiomayServerC() {
    startSiomayServer();
}

extern "C" void sendStopCommandC(int ip[4]) {
    sendStopCommand(ip);
}