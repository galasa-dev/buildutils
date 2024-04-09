package dev.galasa.openapi2beans.example;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.Test;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;

import dev.galasa.openapi2beans.example.generated.BeanWithMultiplePrimitiveProperties;


public class TestBeanWithMultiplePrimitiveProperties {
    
    @Test
    public void TestCanSerialiseTheBean() throws Exception {
        BeanWithMultiplePrimitiveProperties beanUnderTest = new BeanWithMultiplePrimitiveProperties();
        beanUnderTest.SetAStringVariable("hello");
        beanUnderTest.SetAIntVariable(11);
        Gson gson = new GsonBuilder().setPrettyPrinting().create();
        String serialisedForm = gson.toJson(beanUnderTest);
        assertThat(serialisedForm).contains("\"aStringVariable\": \"hello\"");
        assertThat(serialisedForm).contains("\"aIntVariable\": 11");
    }
}
