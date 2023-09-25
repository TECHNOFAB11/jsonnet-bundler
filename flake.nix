{
  description = "A jsonnet package manager";
  inputs.nixpkgs.url = "nixpkgs/nixos-22.11";

  outputs = { self, nixpkgs, systems }:
  let
    version = "0.5.2";
    forEachSystem = nixpkgs.lib.genAttrs (import systems);
    nixpkgsFor = forEachSystem (system: import nixpkgs { inherit system; });
  in {
    # Provide some binary packages for selected system types.
    packages = forEachSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        jb = pkgs.buildGoModule {
          pname = "jsonnet-bundler";
          inherit version;

          src = ./.;
          CGO_ENABLED = 0;

          ldflags = [ "-s" "-w" "-X main.Version=${version}" ];

          vendorSha256 = "sha256-5a+8emtKjOlSr4V9P2YW5u0FVy9to5rIVfF9i90BCe4=";

          doCheck = false;
        };
      }
    );

    defaultPackage = forEachSystem (system: self.packages.${system}.jb);
  };
}
