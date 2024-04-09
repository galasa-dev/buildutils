package dev.galasa.openapi2beans.example.generated;

// a bean with an array referencing an array
public class BeanWith2DReferencedArray {
    // Class Variables //
    // an array variable to be referenced by an array
    private String[][] anArrayVariable;

    // Constants //

    public BeanWith2DReferencedArray () {
    }

    // Getters //
    public String[][] GetAnArrayVariable() {
        return this.anArrayVariable;
    }

    // Setters //
    public void SetAnArrayVariable(String[][] anArrayVariable) {
        this.anArrayVariable = anArrayVariable;
    }
}