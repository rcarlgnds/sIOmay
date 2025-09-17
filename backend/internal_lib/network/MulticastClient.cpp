#include "Multicast.hpp"
void  UdpMultiCastClient::listen_loop() {
    std::array<uint8_t, 1024> recv_buffer;
    asio::ip::udp::endpoint sender_endpoint;

    MouseState mState;
    MouseState prevMState;

    KeyboardState kState;
    while (true) {
        try {
            size_t bytes_received = socket.receive_from(
                asio::buffer(recv_buffer), sender_endpoint);

            if (bytes_received != 16) { // check if buffer is valid
              std::cout<<"Catch data with unrecognize bytesize"<<std::endl;
              continue;
            }

            auto data = recv_buffer.data();
            if (isMouseData(data, 16)){
                parseMouseData(mState, data, bytes_received);
                  // Handle the received MouseState
                //   std::cout << "Received packet from " 
                //               << sender_endpoint.address().to_string() << ": "
                //               << "dx=" << mState.dx 
                //               << ", dy=" << mState.dy 
                //               << ", left=" << mState.leftClick
                //               << ", right=" << mState.rightClick
                //               << ", mid=" << mState.midClick
                //               << std::endl;
                WinApplyMouseState(mState, prevMState);
            }else if (isKeyboardData(data, 16)){
                parseKeyboardData(kState, data, 16);
                // std::cout
                //     << "isKeyDown=" << kState.press 
                //     << ", code=" << kState.code
                //     << std::endl;
                WinApplyKeyInput(kState.press, kState.code );
            }else if (isCommandData(data, 16)){
                int ipdata[4]= {0};
                SystemCommand cmd = parseCommandData(ipdata, data, 16);

                if (cmd == SystemCommand::STOP){
                    // check if this cliennt has that target ip
                    if (has_ip(ipdata)){
                        return;
                    }
                }
            }
            
        } catch (std::exception& e) {
            std::cerr << "Receive error: " << e.what() << std::endl;
        }
    }
}
