# g-tui-go

A terminal UI for interacting with Google's Gemini models.

## Installation

### Nix / NixOS

If you are a Nix user, you can install `g-tui-go` directly from this GitHub repository.

**Option 1: Temporary Shell**

This is the quickest way to run the application without installing it permanently.

1.  **Using `nix-shell`:**
    
    ```bash
    nix-shell -p "git; pkgs.callPackage ./. {}"
    # Inside the new shell, run the application
    g-tui-go
    ```

2.  **Using `nix run` (Flakes required):**
    
    If you have flakes enabled, you can run it directly:
    
    ```bash
    nix run github:garthtrickett/g-tui-go
    ```

**Option 2: Declarative Install (for NixOS or Home Manager)**

To install the application permanently and declaratively, add it to your system or user configuration.

1.  **Add the repository to your inputs** (e.g., in your `flake.nix`):
    
    ```nix
    # /etc/nixos/flake.nix or ~/.config/home-manager/flake.nix
    {
      inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
        
        # Add this line
        g-tui-go = {
          url = "github:garthtrickett/g-tui-go";
          flake = false; # This is not a flake... yet
        };
      };
    
      outputs = { self, nixpkgs, g-tui-go }: {
        # ... your configuration
      };
    }
    ```

2.  **Add the package to your environment:**
    
    Create an overlay to make the package from the `default.nix` file available.
    
    ```nix
    # In your NixOS or Home Manager configuration
    
    let
      # Define your g-tui-go package
      g-tui-go-pkg = pkgs.callPackage inputs.g-tui-go { };
    in
    {
      # For NixOS:
      environment.systemPackages = with pkgs; [
        g-tui-go-pkg
      ];
    
      # OR
    
      # For Home Manager:
      home.packages = with pkgs; [
        g-tui-go-pkg
      ];
    }
    ```

3.  **Rebuild your configuration:**
    
    ```bash
    # For NixOS
    sudo nixos-rebuild switch --flake .
    
    # For Home Manager
    home-manager switch --flake .
    ```

Now, `g-tui-go` will be available as a command in your shell.