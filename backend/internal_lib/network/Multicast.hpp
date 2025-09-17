#pragma once
#include "../../vendor/asio/asio/include/asio.hpp"
#include "../mouse/Mouse.hpp"  
#include "../keyboard/keyboard.hpp"
#include "../tools/serializer.hpp"
#include "../mouse/MouseTracker.hpp"

#include <iostream>
#include <thread>
#include <chrono>
#include <string>

struct UdpMulticastServer {
    asio::io_context& io_context;
    asio::ip::udp::socket socket;
    asio::ip::udp::endpoint multicast_endpoint;

    UdpMulticastServer(asio::io_context& io, const std::string& address,
                       unsigned short port)
        : io_context(io),
          socket(io_context, asio::ip::udp::v4()),
          multicast_endpoint(asio::ip::make_address(address), port)
    {

    }

    void send_loop(int interval_seconds, MouseCapture * mouseCapture);
    void send_loop(int interval_seconds, KeyboardCapture * KeyboardCapture);
    void send_command(uint8_t *data, int len);
};


struct UdpMultiCastClient{
    asio::io_context& io_context;
    asio::ip::udp::socket socket;
    asio::ip::udp::endpoint listen_endpoint;
    asio::ip::address multicast_address;

    UdpMultiCastClient(asio::io_context &io, std::string& multicast_ip, int port ):
        io_context(io),
        socket(io_context, asio::ip::udp::v4()),
        listen_endpoint(asio::ip::udp::v4(), port),
        multicast_address(asio::ip::make_address(multicast_ip)){

        // Allow multiple clients to bind to the same port
        socket.set_option(asio::ip::udp::socket::reuse_address(true));
        std::cout<<"try to bind:"<<multicast_address<<std::endl;
        socket.bind(listen_endpoint);

        // Join multicast group
        socket.set_option(asio::ip::multicast::join_group(multicast_address));
    }

     void listen_loop();   
};

std::string getHostname();


bool has_ip(int ip_arr[4]);