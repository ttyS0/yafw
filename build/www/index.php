<?php
    $hostname = gethostname();
    $server_ip = $_SERVER['SERVER_ADDR'];
    $client_ip = $_SERVER['REMOTE_ADDR'];

    echo <<< EOF
    Server Name: $hostname
    Server IP: $server_ip

    Client IP: $client_ip

    EOF;