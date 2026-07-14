package com.example.orders.main;

import com.example.orders.api.OrderCommandService;
import com.example.orders.app.OrderCommandServiceImpl;
import com.example.orders.base.InMemoryOrderRepository;
import com.example.orders.domain.OrderRepository;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/** main 负责依赖组装；应用层和基础设施层不互相依赖。 */
@Configuration
class OrderSystemConfiguration {

    @Bean
    OrderRepository orderRepository() {
        return new InMemoryOrderRepository();
    }

    @Bean
    OrderCommandService orderCommandService(OrderRepository orderRepository) {
        return new OrderCommandServiceImpl(orderRepository);
    }
}
