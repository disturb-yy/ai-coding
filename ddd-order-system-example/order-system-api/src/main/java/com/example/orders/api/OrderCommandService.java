package com.example.orders.api;

public interface OrderCommandService {

    OrderView create(CreateOrderCommand command);

    OrderView confirm(String orderId);

    OrderView get(String orderId);
}
