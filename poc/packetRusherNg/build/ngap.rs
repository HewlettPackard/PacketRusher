/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
use std::fs;

use asn1_compiler::{
    generator::{Codec, Derive, Visibility},
    Asn1Compiler,
};

pub fn generate_ngap() -> std::io::Result<()> {
    let files: Vec<_> = fs::read_dir("ngap").unwrap().map(|f|f.unwrap().path()).collect();

    let mut compiler = Asn1Compiler::new(
        "src/ngap.rs",
        &Visibility::Public,
        vec![Codec::Aper],
        vec![
            Derive::Debug,
            Derive::EqPartialEq,
            Derive::Serialize,
            Derive::Deserialize,
        ],
    );
    compiler.compile_files(&files)?;

    Ok(())
}
