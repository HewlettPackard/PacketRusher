/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
mod ngap;
mod nas;
use std::error::Error;

fn main() -> Result<(), Box<dyn Error>> {
    ngap::generate_ngap()?;
    nas::generate_nas()?;

    Ok(())
}
