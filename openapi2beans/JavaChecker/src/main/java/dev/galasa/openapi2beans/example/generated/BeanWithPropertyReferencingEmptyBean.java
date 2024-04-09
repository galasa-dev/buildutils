package dev.galasa.openapi2beans.example.generated;

// A bean to test referencing functionality
public class BeanWithPropertyReferencingEmptyBean {
    // Class Variables //
    // An empty bean with no properties
    private EmptyBean referencingProperty;

    // Constants //

    public BeanWithPropertyReferencingEmptyBean () {
    }

    // Getters //
    public EmptyBean GetReferencingProperty() {
        return this.referencingProperty;
    }

    // Setters //
    public void SetReferencingProperty(EmptyBean referencingProperty) {
        this.referencingProperty = referencingProperty;
    }
}