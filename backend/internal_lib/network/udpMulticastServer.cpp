#include "Multicast.hpp"

// INI SANGAT GK BAGUS, BUAT KEJAR DEADLINE, DI OVERLOAD AJA DULU,
// NANTI VU BENERIN BIAR BAGUS, MAU PAKE CALLBACK ATAU APA TERSERAH

void UdpMulticastServer::send_loop(int interval_seconds, MouseCapture * mouseCapture) {
    MouseState state;
    uint8_t buf[16]={};
    while (true) {
        if (mouseCapture->poll(state)){
            formatMouseData(state, buf, 16);
            // std::cout
            //         << "dx=" << state.dx 
            //         << ", dy=" << state.dy 
            //         << ", dz=" << state.dScroll 
            //         << ", left=" << state.leftClick
            //         << ", right=" << state.rightClick
            //         << ", mid=" << state.midClick
            //         << std::endl;
            socket.send_to(asio::buffer(buf, 16), multicast_endpoint);
        }
        
        std::this_thread::sleep_for(std::chrono::milliseconds(interval_seconds));
    }
}

void UdpMulticastServer::send_loop(int interval_seconds, KeyboardCapture * KeyboardCapture){
    KeyboardState state;
    uint8_t buf[16]={};
    while (true) {
        // std::cout<<"looping"<<std::endl;
        if (KeyboardCapture->poll(state)){
            formatKeyboardData(state, buf, 16);
            std::cout
                    << "isKeyDown=" << state.press 
                    << ", code=" << state.code
                    << std::endl;
            socket.send_to(asio::buffer(buf, 16), multicast_endpoint);
        }
        
        std::this_thread::sleep_for(std::chrono::milliseconds(interval_seconds));
    }
}

void UdpMulticastServer::send_command(uint8_t *data, int len){
    socket.send_to(asio::buffer(data, len), multicast_endpoint);
}