package dev.galasa.openapi2beans.example.generated;

// A bean with a required property referencing another object
public class BeanWithRequiredReferencedObject {
    // Class Variables //
    // An empty bean with no properties
    private EmptyBean referencingObject;

    // Constants //

    public BeanWithRequiredReferencedObject () {
    }

    // Getters //
    public EmptyBean GetReferencingObject() {
        return this.referencingObject;
    }

    // Setters //
    public void SetReferencingObject(EmptyBean referencingObject) {
        this.referencingObject = referencingObject;
    }
}