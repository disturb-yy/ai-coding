package com.example.orders.domain;

import java.util.Objects;

/** 订单聚合根：所有状态转换均由聚合行为保护。 */
public final class Order {

    private final OrderId id;
    private final String customerId;
    private final Money total;
    private OrderStatus status;
    private String cancelReason;

    private Order(OrderId id, String customerId, Money total, OrderStatus status, String cancelReason) {
        this.id = Objects.requireNonNull(id, "订单 ID 不能为空");
        this.customerId = requireText(customerId, "客户 ID 不能为空");
        this.total = Objects.requireNonNull(total, "订单金额不能为空");
        this.status = Objects.requireNonNull(status, "订单状态不能为空");
        this.cancelReason = cancelReason;
    }

    public static Order create(OrderId id, String customerId, Money total) {
        return new Order(id, customerId, total, OrderStatus.PENDING, null);
    }

    public void confirm() {
        if (status != OrderStatus.PENDING) {
            throw new IllegalStateException("只有待确认订单可以确认");
        }
        status = OrderStatus.CONFIRMED;
    }

    public void cancel(String reason) {
        if (status == OrderStatus.CONFIRMED) {
            throw new IllegalStateException("已确认订单不能取消");
        }
        if (status == OrderStatus.CANCELLED) {
            throw new IllegalStateException("订单已经取消");
        }
        cancelReason = requireText(reason, "取消原因不能为空");
        status = OrderStatus.CANCELLED;
    }

    public OrderId id() { return id; }
    public String customerId() { return customerId; }
    public Money total() { return total; }
    public OrderStatus status() { return status; }
    public String cancelReason() { return cancelReason; }

    private static String requireText(String value, String message) {
        if (value == null || value.isBlank()) {
            throw new IllegalArgumentException(message);
        }
        return value;
    }
}
