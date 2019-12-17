package com.huajiao;

import com.huajiao.common.client.anno.EnablePepperBus;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@EnablePepperBus
@SpringBootApplication
public class PepperbusDemoApplication {

	public static void main(String[] args) {
		SpringApplication.run(PepperbusDemoApplication.class, args);
	}
}
