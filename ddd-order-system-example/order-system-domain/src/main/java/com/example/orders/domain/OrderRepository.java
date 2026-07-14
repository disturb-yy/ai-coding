package com.example.orders.domain;

import java.util.Optional;

/** 领域定义仓储能力，基础设施模块负责实现。 */
public interface OrderRepository {

    Optional<Order> findById(OrderId id);

    void save(Order order);
}
