package dev.galasa.openapi2beans.example.generated;

// a bean with a reference to a property in the wider schema map
public class BeanWithReferenceToProp {
    // Class Variables //
    // an array variable to be referenced by an array
    private String[] aReferencingVar;

    // Constants //

    public BeanWithReferenceToProp () {
    }

    // Getters //
    public String[] GetAReferencingVar() {
        return this.aReferencingVar;
    }

    // Setters //
    public void SetAReferencingVar(String[] aReferencingVar) {
        this.aReferencingVar = aReferencingVar;
    }
}