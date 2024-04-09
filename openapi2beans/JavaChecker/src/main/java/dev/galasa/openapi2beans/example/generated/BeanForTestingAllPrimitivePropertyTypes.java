package dev.galasa.openapi2beans.example.generated;

// A bean that tests all primitive property types (except arrays)
public class BeanForTestingAllPrimitivePropertyTypes {
    // Class Variables //
    private boolean aBooleanVariable;
    // this should be an int in a java bean
    private int aIntVariable;
    // this should be a float in the java bean
    private double aNumberVariable;
    // this should be a String in a java bean
    private String aStringVariable;

    // Constants //

    public BeanForTestingAllPrimitivePropertyTypes () {
    }

    // Getters //
    public boolean GetABooleanVariable() {
        return this.aBooleanVariable;
    }
    public int GetAIntVariable() {
        return this.aIntVariable;
    }
    public double GetANumberVariable() {
        return this.aNumberVariable;
    }
    public String GetAStringVariable() {
        return this.aStringVariable;
    }

    // Setters //
    public void SetABooleanVariable(boolean aBooleanVariable) {
        this.aBooleanVariable = aBooleanVariable;
    }
    public void SetAIntVariable(int aIntVariable) {
        this.aIntVariable = aIntVariable;
    }
    public void SetANumberVariable(double aNumberVariable) {
        this.aNumberVariable = aNumberVariable;
    }
    public void SetAStringVariable(String aStringVariable) {
        this.aStringVariable = aStringVariable;
    }
}