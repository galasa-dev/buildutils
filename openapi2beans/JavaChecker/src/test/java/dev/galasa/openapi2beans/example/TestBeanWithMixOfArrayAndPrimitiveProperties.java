package dev.galasa.openapi2beans.example;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.Test;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;

import dev.galasa.openapi2beans.example.generated.BeanWithMixOfArrayAndPrimitiveProperties;


public class TestBeanWithMixOfArrayAndPrimitiveProperties {
    
    
    @Test
    public void TestCanSerialiseTheBean() throws Exception {
        BeanWithMixOfArrayAndPrimitiveProperties beanUnderTest = new BeanWithMixOfArrayAndPrimitiveProperties();
        beanUnderTest.SetAStringVariable("hello");
        beanUnderTest.SetAnIntVariable(11);
        beanUnderTest.SetAnArrayVariable(new String[]{"string0", "string1"});
        beanUnderTest.SetAnIntArray(new int[]{1, 2});
        Gson gson = new GsonBuilder().setPrettyPrinting().create();
        String serialisedForm = gson.toJson(beanUnderTest);
        assertThat(serialisedForm).contains("\"aStringVariable\": \"hello\"");
        assertThat(serialisedForm).contains("\"anIntVariable\": 11");
        assertThat(serialisedForm).contains("\"anIntArray\": [\n" +
                                            "    1,\n" +
                                            "    2\n" + 
                                            "  ]");
        assertThat(serialisedForm).contains("\"anArrayVariable\": [\n" +
                                            "    \"string0\",\n" +
                                            "    \"string1\"\n" +
                                            "  ]");
    }
}
