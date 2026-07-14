package com.example.orders.app;

import com.example.orders.api.CreateOrderCommand;
import com.example.orders.api.OrderCommandService;
import com.example.orders.api.OrderView;
import com.example.orders.domain.Money;
import com.example.orders.domain.Order;
import com.example.orders.domain.OrderId;
import com.example.orders.domain.OrderRepository;
import java.util.Objects;

/** 应用层只编排用例；状态变化委托给订单聚合。 */
public final class OrderCommandServiceImpl implements OrderCommandService {

    private final OrderRepository orderRepository;

    public OrderCommandServiceImpl(OrderRepository orderRepository) {
        this.orderRepository = Objects.requireNonNull(orderRepository);
    }

    @Override
    public OrderView create(CreateOrderCommand command) {
        Objects.requireNonNull(command, "创建订单命令不能为空");
        var order = Order.create(OrderId.newId(), command.customerId(), new Money(command.totalAmount()));
        orderRepository.save(order);
        return toView(order);
    }

    @Override
    public OrderView confirm(String orderId) {
        var order = load(orderId);
        order.confirm();
        orderRepository.save(order);
        return toView(order);
    }

    @Override
    public OrderView get(String orderId) {
        return toView(load(orderId));
    }

    private Order load(String orderId) {
        return orderRepository.findById(OrderId.from(orderId))
                .orElseThrow(() -> new IllegalArgumentException("订单不存在: " + orderId));
    }

    private OrderView toView(Order order) {
        return new OrderView(order.id().toString(), order.customerId(), order.total().amount(), order.status().name());
    }
}
