package dev.galasa.openapi2beans.example.generated;

// A bean with a single required primitive property
public class BeanWithRequiredPrimitiveProperty {
    // Class Variables //
    private String aStringVariable;

    // Constants //

    public BeanWithRequiredPrimitiveProperty (String aStringVariable) {
        this.aStringVariable = aStringVariable;
    }

    // Getters //
    public String GetAStringVariable() {
        return this.aStringVariable;
    }

    // Setters //
    public void SetAStringVariable(String aStringVariable) {
        this.aStringVariable = aStringVariable;
    }
}